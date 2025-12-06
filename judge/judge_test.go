package judge

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/containers/podman/v5/pkg/bindings/containers"
)

func tryConnect(t *testing.T) context.Context {
	if os.Getenv("PODMAN_URI") == ""{
		t.Skip("Skipping test, no URI set")
	}

	conn, err := ConnectToPodman(os.Getenv("PODMAN_URI"))
	if err != nil {
		t.Skip("Skipping test, Podman connection failed")
	}

	return conn
}

func test(t *testing.T, program, input, output, expectedMatch string) {
	conn := tryConnect(t)

	createResponse, err := CreateSandbox(conn, program, input, output)
	if err != nil {
		t.Errorf("Could not create sandbox: %v\n", err)
		return
	}
	t.Logf("%v\n", err)

	stdoutBuf := new(strings.Builder)
	err = StartSandbox(conn, createResponse, stdoutBuf)
	if err != nil {
		t.Errorf("Could not start sandbox: %v", err)
		return
	}

	_, err = containers.Wait(conn, createResponse.ID, nil)
	if err != nil {
		t.Errorf("Could not stop sandbox: %v", err)
		return
	}

	resp := stdoutBuf.String()
	resp = strings.TrimSpace(resp)
	if match, _ := regexp.MatchString(expectedMatch, resp); !match {
		t.Errorf("Expected %v, got: '%v'\n", expectedMatch, resp)
	}
}

func TestTLE(t *testing.T) {
	program := `
package main

import (
	"fmt"
	"time"
)

func main() {
	time.Sleep(4 * time.Second)
	fmt.Println(5 * 2)
}
`
	input := "10\n"
	output := "10\n"

	test(t, program, input, output, "Time limit exceeded")
}

func TestWA(t *testing.T) {
	program := `
package main

import "fmt"

func main() {
	fmt.Println(5 * 2 + 1)
}
`
	input := "10\n"
	output := "10\n"

	test(t, program, input, output, "Wrong answer")
}

func TestRE(t *testing.T) {
	program := `
package main

import "fmt"

func main() {
	zero := 0
	fmt.Println(5 * 2 + 1 / zero)
}
`
	input := "10\n"
	output := "10\n"

	test(t, program, input, output, "Runtime error")
}

func TestBE(t *testing.T) {
	program := `
package main

import "fmt"

func main( {
	fmt.Println(5 * 2)
}
`
	input := "10\n"
	output := "10\n"

	test(t, program, input, output, "^[0-9]\\.[0-9]+s$")
}

func TestPass(t *testing.T) {
	program := `
package main

import "fmt"

func main() {
	fmt.Println(5 * 2)
}
`
	input := "10\n"
	output := "10\n"

	test(t, program, input, output, "^[0-9]\\.[0-9]+s$")
}

