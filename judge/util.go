package judge

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
)

type SubmissionTestCase struct {
	Input     string
	Output    string
}

type Worker struct {
	Uri       string
	Ctx       context.Context
	Available bool
}

func genRandStr(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

func writeStringToFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	f.Sync()
	fmt.Println("Cerrando: ", path)

	return nil
}
