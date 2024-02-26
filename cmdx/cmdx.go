package cmdx

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func ExecShellCmd(cmdStr string, pctx context.Context) error {
	return ExecShellCmdWithOutput(cmdStr, pctx, os.Stdout)
}

func ExecShellCmdWithOutput(cmdStr string, pCtx context.Context, output io.Writer) error {
	ctx := context.WithoutCancel(pCtx)

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdStr)

	return doExecCmd(cmd, output)
}

func ExecBinaryCmd(cmdStr string, pCtx context.Context) error {
	if len(cmdStr) == 0 {
		return nil
	}

	ctx := context.WithoutCancel(pCtx)
	parts := strings.Fields(cmdStr)
	head := parts[0]
	parts = parts[1:]
	cmd := exec.CommandContext(ctx, head, parts...)

	return doExecCmd(cmd, os.Stdout)
}

// TODO remove debug logs
func doExecCmd(cmd *exec.Cmd, output io.Writer) (err error) {
	cmd.Stdout = output
	cmd.Stderr = output
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	if cmd.ProcessState != nil && cmd.Process != nil {
		ret := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		if ret != 0 {
			return fmt.Errorf("Exit with error code:%v", ret)
		}
	}

	return nil
}
