package cmdx

import (
	"context"
	"io"
	"os/exec"
	"strings"

	"github.com/madlabx/pkgx/errors"
)

type Result struct {
	Stdout   io.Writer
	Stderr   io.Writer
	ExitCode int
}

func ExecShellCmd(pCtx context.Context, cmdStr string, result *Result) error {
	if len(cmdStr) == 0 {
		return errors.New("empty cmdStr")
	}
	ctx := context.WithoutCancel(pCtx)
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdStr)

	return doExecCmd(cmd, result)
}

func ExecBinaryCmd(pCtx context.Context, cmdStr string, result *Result) error {
	if len(cmdStr) == 0 {
		return errors.New("empty cmdStr")
	}

	ctx := context.WithoutCancel(pCtx)
	parts := strings.Fields(cmdStr)
	head := parts[0]
	parts = parts[1:]
	cmd := exec.CommandContext(ctx, head, parts...)

	return doExecCmd(cmd, result)
}

func doExecCmd(cmd *exec.Cmd, cr *Result) error {
	if cr != nil {
		cmd.Stdout = cr.Stdout
		cmd.Stderr = cr.Stderr
	}

	err := cmd.Run()

	if cr != nil {
		cr.ExitCode = cmd.ProcessState.ExitCode()
	}

	return errors.Wrap(err)
}
