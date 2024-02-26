package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/madlabx/pkgx/cmdx"
)

func main() {
	cmdStr := strings.Join(os.Args[1:], " ")
	err := cmdx.ExecBinaryCmd(cmdStr, context.Background())
	if err != nil {
		fmt.Println(err)
	}

}
