package judge

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"

	"log"
	"os"

	buildahDefine "github.com/containers/buildah/define"
	"github.com/containers/podman/v5/pkg/api/handlers"
	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/domain/entities/types"
	"github.com/containers/podman/v5/pkg/specgen"
	dockerContainer "github.com/docker/docker/api/types/container"
	"lidsol.org/papeador/store"
	_ "modernc.org/sqlite"
)

var podmanConns []*Worker
var workerQueue = make(chan *Worker, 10)

var WorkerQueueP = &workerQueue

var m sync.Mutex

func ConnectToPodman(connURI string) (context.Context, error) {
	conn, err := bindings.NewConnection(context.Background(), connURI)
	if err != nil {
		return nil, err
	}

	worker := &Worker{
		Uri:       connURI,
		Ctx:       conn,
		Available: true,
	}
	podmanConns = append(podmanConns, worker)

	workerQueue <- worker

	log.Println("Connection successful")

	return conn, nil
}

func CreateFiles(filetype, programStr string, testcases []SubmissionTestCase, timeLimit int) error {
	m.Lock()
	defer m.Unlock()

	err := os.MkdirAll("/vol/podman/inputs", 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll("/vol/podman/expected-outputs", 0755)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	defer func() {
		err = os.Chdir(cwd)
	}()

	err = os.Chdir("/vol/podman")
	if err != nil {
		return err
	}

	programPath := fmt.Sprintf("./papeador-submission.%v", filetype)

	timeLimitPath := "/vol/podman/timelimit.txt"
	timeLimitString := strconv.Itoa(timeLimit) + "\n"
	err = writeStringToFile(timeLimitPath, timeLimitString)
	if err != nil {
		return err
	}

	err = writeStringToFile(programPath, programStr)
	if err != nil {
		return err
	}

	for k, testcase := range testcases {
		testInputPath := fmt.Sprintf("./inputs/%02d.txt", k)
		expectedOutputPath := fmt.Sprintf("./expected-outputs/%02d.txt", k)

		err = writeStringToFile(testInputPath, testcase.Input)
		if err != nil {
			return err
		}

		err = writeStringToFile(expectedOutputPath, testcase.Output)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateSandbox(conn context.Context, filetype, programStr string, testcases []SubmissionTestCase, timeLimit int, output int) (types.ContainerCreateResponse, error) {

	err := CreateFiles(filetype, programStr, testcases, timeLimit)

	options := types.BuildOptions{
		BuildOptions: buildahDefine.BuildOptions{
			Output:           "program:latest",
			ConfigureNetwork: buildahDefine.NetworkDisabled,
		},
	}

	log.Println("Building image")

	dockerfilePath := fmt.Sprintf("/vol/podman/Dockerfile-%v", filetype)
	log.Println(dockerfilePath)
	_, err = images.Build(conn, []string{dockerfilePath}, options)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	log.Println("Creating with spec")
	s := specgen.NewSpecGenerator("program:latest", false)
	s.Command = []string{"/bin/alive"}
	s.Name = "submission-sandbox" + genRandStr(8)
	createReponse, err := containers.CreateWithSpec(conn, s, nil)
	if err != nil {
		return types.ContainerCreateResponse{}, err
	}

	return createReponse, nil
}

func StartSandbox(conn context.Context, createResponse types.ContainerCreateResponse, w io.Writer) error {
	log.Println(createResponse.ID)
	if err := containers.Start(conn, createResponse.ID, &containers.StartOptions{}); err != nil {
		return err
	}
	dockerExecOpts := dockerContainer.ExecOptions{
		Cmd:          []string{"/bin/papeador"},
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  false,
		Tty:          true,
	}
	execConfig := &handlers.ExecCreateConfig{dockerExecOpts}
	execId, err := containers.ExecCreate(conn, createResponse.ID, execConfig)
	if err != nil {
		return err
	}
	_ = execId

	var stderr io.Writer = os.Stdout
	input := bufio.NewReader(os.Stdin)
	_ = input

	attachOptions := new(containers.ExecStartAndAttachOptions)
	attachOptions.WithAttachError(true).WithAttachInput(false).WithAttachOutput(true).WithErrorStream(stderr).WithOutputStream(w)
	err = containers.ExecStartAndAttach(conn, execId, attachOptions)

	if err != nil {
		return err
	}

	return nil
}

func CalculateAndUpdateRanking(
	ctx context.Context,
	s store.Store,
	userID int,
	status string,
	executionTime float64) ([]store.UserScore, error) {

	log.Printf("Iniciando cálculo de puntaje para el usuario: %d, status: %s", userID, status)
	//calculate user score
	input := store.ScoringInput{Status: status, ExecutionTime: executionTime}
	score := calculateSingleScore(input)

	log.Printf("Puntaje calculado: %d", score)

	//funcion para cambiar el puntaje del usuario
	if err := s.UserScore(ctx, userID, score); err != nil {
		log.Printf("Error al actualizar puntaje del usuario %d: %v", userID, err)
		return nil, fmt.Errorf("error al actualizar puntaje en store: %w", err)
	}

	ranking, err := s.GetRanking(ctx)
	if err != nil {
		log.Printf("Error al obtener ranking: %v", err)
		return nil, fmt.Errorf("error al obtener ranking del store: %w", err)
	}

	PrintRanking(ranking)
	return ranking, nil
}

// ---------------------------------------------------------
// funcion adicional para imprimir el ranking
func PrintRanking(ranking []store.UserScore) {
	fmt.Println("\n==============================================")
	fmt.Println(" RANKING DE USUARIOS ")
	fmt.Println("----------------------------------------------")
	fmt.Printf("%-5s | %-15s | %s\n", "RANK", "USER", "POINTS")
	fmt.Println("----------------------------------------------")

	for _, user := range ranking {
		fmt.Printf("%-5d | %-15d | %d\n", user.Rank, user.UserID, user.Score)
	}
	fmt.Println("==============================================")
}

// funcion provicional para calcular el puntaje de un solo usuario
func calculateSingleScore(sub store.ScoringInput) int {
	return 0
}
