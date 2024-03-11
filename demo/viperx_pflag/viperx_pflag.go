package main

import (
	"fmt"
	"os"

	"github.com/madlabx/pkgx/viperx"
	"github.com/spf13/pflag"
)

var cfgFile string

type ConfigSys struct {
	LogLevel string
	Size     int
}

type UserCfg struct {
	LogLevel string
	Size     int
}

type NetCap struct {
	Bw int64 `vx_flag:";;100;bandwith"`
}

type Config struct {
	Sys ConfigSys
	Ttt int64 `vx_name:"ttt" vx_short:"t" vx_default:"1234" vx_desc:"test for ttt"`
	Nc  NetCap
}

var viperxConif Config

func main() {
	// Here we define our flags, and bind them to viper configurations
	pflag.StringVar(&cfgFile, "config", "", "config file")

	pflag.String("sys.loglevel", "default_flag_level", "loglevel")

	if _, err := viperx.BindAllFlags(pflag.CommandLine, Config{}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pflag.Parse()

	if err := viperx.ParseConfig(&viperxConif, "DEMO", "../conf/viperx.jsonx"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Config is: %#v\n", viperxConif)

}
