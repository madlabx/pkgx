package cmdx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindowsExec(t *testing.T) {
	err := ExecBinaryCmd(context.Background(), "cmd /c type ..\\demo\\cmdx\\cmdx.go", nil)
	assert.Nil(t, err)
	//cmd := exec.Command("cmd", "/c", `type ..\demo\cmdx\cmdx.go`)
	//// 使用 PowerShell 命令
	////cmd := exec.Command("powershell", "Get-Content file.txt")
	//
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//cmd.Stderr = &out
	//err := cmd.Run()
	//if err != nil {
	//	fmt.Printf("Error executing command:%v, out:%v", err, out.String())
	//	return
	//}
	//
	//// 打印命令的输出
	//fmt.Print(out.String())
}
