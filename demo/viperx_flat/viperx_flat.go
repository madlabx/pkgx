package main

import (
	"fmt"
	"os"

	"github.com/madlabx/pkgx/viperx"
	"github.com/spf13/pflag"
)

func main() {
	var cfgFile string
	pflag.StringVar(&cfgFile, "config", "../conf/viperx.json", "config file")
	pflag.String("sys.loglevel", "KO", "log level")

	if err := viperx.BindPFlags(pflag.CommandLine); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pflag.Parse()

	viperx.SetConfigFile(cfgFile)
	if err := viperx.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Config is: %#v\n", viperx.ConfigFileUsed())

	sysLogLevel := viperx.GetString("sys.loglevel", "KO")
	fmt.Printf("sysLogLevel:%v\n", sysLogLevel)

	attrNotDefined := viperx.GetString("sys.attrNotDefined", "OK")
	fmt.Printf("attrNotDefined:%v\n", attrNotDefined)
}
