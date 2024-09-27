package cmdx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
)

type Output struct {
	Stdout io.Writer
	Stderr io.Writer
}

func ExecShellCmd(pCtx context.Context, cmdStr string, result *Output) error {
	if len(cmdStr) == 0 {
		return ErrEmptyCmdStr
	}
	ctx := context.WithoutCancel(pCtx)
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	return doExecCmd(cmd, result)
}

func ExecBinaryCmd(pCtx context.Context, cmdStr string, result *Output) error {
	if len(cmdStr) == 0 {
		return ErrEmptyCmdStr
	}

	ctx := context.WithoutCancel(pCtx)
	parts := strings.Fields(cmdStr)
	head := parts[0]
	parts = parts[1:]
	cmd := exec.CommandContext(ctx, head, parts...)

	return doExecCmd(cmd, result)
}
func ShExecWithResponse(pCtx context.Context, cmdStr string, resp any) error {
	return doExecWithResponse(true, pCtx, cmdStr, resp)
}
func ExecWithResponse(pCtx context.Context, cmdStr string, resp any) error {
	return doExecWithResponse(false, pCtx, cmdStr, resp)
}

func doExecWithResponse(UnderShell bool, pCtx context.Context, cmdStr string, resp any) error {
	workerFn := ExecBinaryCmd
	if UnderShell {
		workerFn = ExecShellCmd
	}
	if len(cmdStr) == 0 {
		return ErrEmptyCmdStr
	}
	var (
		op    Output
		begin = time.Now()
	)
	defer func() {
		log.Infof("ExecWithResponse: %s, cost: %s", cmdStr, time.Since(begin))
	}()
	if resp == nil {
		op = Output{
			Stdout: log.StandardLogger().Out,
			Stderr: log.StandardLogger().Out,
		}
		if err := workerFn(pCtx, cmdStr, &op); err != nil {
			log.Errorf("Failed to execute [%v], err:%v", cmdStr, err)
			return errors.Wrap(err)
		}
	} else {
		var stdOut, stdErr bytes.Buffer
		op = Output{
			Stdout: &stdOut,
			Stderr: &stdErr,
		}
		if err := workerFn(pCtx, cmdStr, &op); err != nil {
			log.Errorf("Failed to execute [%v], stdout:[%v], stderr:[%v], err:%v", cmdStr, stdOut.String(), stdErr.String(), err)
			return errors.Wrap(err)
		} else {
			if err = json.Unmarshal(stdOut.Bytes(), resp); err != nil {
				log.Errorf("Failed to Unmarshal output of cmdStr:[%v], stdout:[%v], stderr:[%v], err:%v", cmdStr, stdOut.String(), stdErr.String(), err)
				return errors.Wrap(err)
			}
		}
	}
	return nil

}

func doExecCmd(cmd *exec.Cmd, cr *Output) error {
	if cr != nil {
		cmd.Stdout = cr.Stdout
		cmd.Stderr = cr.Stderr
	} else {
		cmd.Stdout = log.StandardLogger().Out
		cmd.Stderr = log.StandardLogger().Out
	}

	return errors.Wrap(cmd.Run())
}
