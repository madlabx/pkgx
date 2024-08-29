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
	err := cmdx.ExecBinaryCmd(context.Background(), cmdStr, nil)
	if err != nil {
		fmt.Println(err)
	}

}
