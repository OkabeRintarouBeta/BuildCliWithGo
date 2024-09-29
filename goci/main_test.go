package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func mockCmdContext(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess"}
	cs = append(cs, exe)
	cs = append(cs, args...)

	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func mockCmdTimeout(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cmd := mockCmdContext(ctx, exe, args...)
	cmd.Env = append(cmd.Env, "GO_HELPER_TIMEOUT=1")
	return cmd
}

func TestRun(t *testing.T) {
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git not installed. Skipping test.")
	}
	testCases := []struct {
		name     string
		proj     string
		out      string
		expErr   error
		setupGit bool
		mockCmd  func(ctx context.Context, name string, arg ...string) *exec.Cmd
	}{
		{
			name: "success",
			proj: "./testdata/tool",
			out: "Go Build: SUCCESS\n" +
				"Go Test: SUCCESS\n" +
				"Gofmt: SUCCESS\n" +
				"Git Push: SUCCESS\n",
			expErr:   nil,
			setupGit: true,
			mockCmd:  nil,
		},
		{
			name: "successMock",
			proj: "./testdata/tool",
			out: "Go Build: SUCCESS\n" +
				"Go Test: SUCCESS\n" +
				"Gofmt: SUCCESS\n" +
				"Git Push: SUCCESS\n",
			expErr:   nil,
			setupGit: false,
			mockCmd:  mockCmdContext,
		},
		{
			name:     "fail",
			proj:     "./testdata/toolErr",
			out:      "",
			expErr:   &stepError{step: "go build"},
			setupGit: false,
			mockCmd:  nil,
		},
		{
			name:     "failFormat",
			proj:     "./testdata/toolFmtErr",
			out:      "",
			expErr:   &stepError{step: "go fmt"},
			setupGit: false,
			mockCmd:  nil,
		},
		{
			name:     "failTimeout",
			proj:     "./testdata/tool",
			out:      "",
			expErr:   context.DeadlineExceeded,
			setupGit: false,
			mockCmd:  mockCmdTimeout,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupGit {
				_, err := exec.LookPath("git")
				if err != nil {
					t.Skip("Git not installed. Skipping test.")
				}
				cleanup := setUpGit(t, tc.proj)
				defer cleanup()
			}

			if tc.mockCmd != nil {
				command = tc.mockCmd
			}

			var out bytes.Buffer
			err := run(tc.proj, &out)

			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected error: %q, got `nil` instead", tc.expErr)
					return
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q, got %q instead", tc.expErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}
			if out.String() != tc.out {
				t.Errorf("Expected output: %q, got %q instead", tc.out, out.String())
			}
		})
	}

}

func setUpGit(t *testing.T, project string) func() {
	t.Helper()
	gitExec, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	tempDir, err := os.MkdirTemp("", "go-ci-test")
	if err != nil {
		t.Fatal(err)
	}
	projectPath, err := filepath.Abs(project)
	if err != nil {
		t.Fatal(err)
	}
	remoteURI := fmt.Sprintf("file://%s", tempDir)
	gitCommandList := []struct {
		args []string
		dir  string
		env  []string
	}{
		{[]string{"init", "--bare"}, tempDir, nil},
		{[]string{"init"}, projectPath, nil},
		{[]string{"remote", "add", "origin", remoteURI}, projectPath, nil},
		{[]string{"add", "."}, projectPath, nil},
		{[]string{"commit", "-m", "test"}, projectPath, []string{
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
		}},
	}
	for _, gitCommand := range gitCommandList {

		command := exec.Command(gitExec, gitCommand.args...)
		command.Dir = gitCommand.dir
		if gitCommand.env != nil {
			command.Env = append(os.Environ(), gitCommand.env...)
		}
		if err := command.Run(); err != nil {
			t.Fatal(err)
		}
	}

	return func() {
		os.RemoveAll(tempDir)
		os.RemoveAll(filepath.Join(projectPath, ".git"))
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("GO_HELPER_TIMEOUT") == "1" {
		time.Sleep(15 * time.Second)
	}
	if os.Args[2] == "git" {
		fmt.Fprintln(os.Stdout, "Everything up-to-date")
		os.Exit(0)
	}
	os.Exit(1)
}

func TestRunKill(t *testing.T) {
	// RunKill Test Cases
	var testCases = []struct {
		name   string
		proj   string
		sig    syscall.Signal
		expErr error
	}{
		{"SIGINT", "./testdata/tool", syscall.SIGINT, ErrSignal},
		{"SIGTERM", "./testdata/tool", syscall.SIGTERM, ErrSignal},
		{"SIGQUIT", "./testdata/tool", syscall.SIGQUIT, nil},
	}

	// RunKill Test Execution
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			command = mockCmdTimeout

			errCh := make(chan error)
			ignSigCh := make(chan os.Signal, 1)
			expSigCh := make(chan os.Signal, 1)

			signal.Notify(ignSigCh, syscall.SIGQUIT)
			defer signal.Stop(ignSigCh)

			signal.Notify(expSigCh, tc.sig)
			defer signal.Stop(expSigCh)

			go func() {
				errCh <- run(tc.proj, ioutil.Discard)
			}()

			go func() {
				time.Sleep(2 * time.Second)
				syscall.Kill(syscall.Getpid(), tc.sig)
			}()

			// select error
			select {
			case err := <-errCh:
				if err == nil {
					t.Errorf("Expected error. Got 'nil' instead.")
					return
				}

				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q. Got %q", tc.expErr, err)
				}

				// select signal
				select {
				case rec := <-expSigCh:
					if rec != tc.sig {
						t.Errorf("Expected signal %q, got %q", tc.sig, rec)
					}
				default:
					t.Errorf("Signal not received")
				}
			case <-ignSigCh:
			}
		})
	}
}
