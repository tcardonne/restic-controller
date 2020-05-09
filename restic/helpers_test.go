package restic

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

// Repositories targets are not used in reality for unit tests
var testResticRepository = "rest:https://user:password@restic.domain.tld/test"
var testResticPassword = "repositoryPassword"

// Mock exec.Command calls
func mockExecOutputString(output string, exitCode int) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"GO_EXIT_CODE=" + strconv.Itoa(exitCode),
			"GO_OUTPUT=" + output,
		}
		return cmd
	}
}

func mockExecOutputFile(outputFile string, exitCode int) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"GO_EXIT_CODE=" + strconv.Itoa(exitCode),
			"GO_OUTPUT_FILE=" + outputFile,
		}
		return cmd
	}
}

func mockExecOutputFileStderr(outputFile string, exitCode int) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"GO_EXIT_CODE=" + strconv.Itoa(exitCode),
			"GO_OUTPUT_FILE=" + outputFile,
			"GO_OUTPUT_STDERR=1",
		}
		return cmd
	}
}

// TestHelperProcess is called by mocked exec.Command calls to simulate Restic invocations.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	exitCode, _ := strconv.Atoi(os.Getenv("GO_EXIT_CODE"))

	outputString := os.Getenv("GO_OUTPUT")
	if len(outputString) > 0 {
		fmt.Fprint(os.Stdout, outputString)
		os.Exit(exitCode)
	}

	outputFile := os.Getenv("GO_OUTPUT_FILE")
	if len(outputFile) > 0 {
		data, err := ioutil.ReadFile(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Loading test file error : %s", err)
			os.Exit(1)
		}
		if os.Getenv("GO_OUTPUT_STDERR") != "1" {
			os.Stdout.Write(data)
		} else {
			os.Stderr.Write(data)
		}
		os.Exit(exitCode)
	}

	os.Exit(exitCode)
}
