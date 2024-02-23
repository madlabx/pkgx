package main

import (
	"fmt"
	"github.com/madlabx/pkgx/viperx"
	"github.com/spf13/pflag"
	"os"
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
	Bw int64 `vxflag:"default:100;desc:bandwith"`
}

type Config struct {
	Sys  ConfigSys
	Ttt  int64 `vxflag:"name:t;default:1234;desc:test for ttt"`
	Nc   NetCap
	User UserCfg
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

	if err := viperx.ParseConfig(&viperxConif, "DEMO", "../conf/viperx.json"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Config is: %#v\n", viperxConif)

}
