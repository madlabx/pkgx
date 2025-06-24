package main

import (
	"fmt"
	"os"

	"github.com/madlabx/pkgx/utils"
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
	Bw int64 //`vx_tag:";;100;bandwith"`
}

type Config struct {
	//Sys   ConfigSys
	//Ttt   int64 `vx_name:"ttt" vx_short:"t" vx_default:"1234" vx_desc:"test for ttt"`
	//Nc    *NetCap
	//User  *UserCfg
	Fam *Name
	//Color *string
}
type Name struct {
	Family string
	Size   int
}

var viperxConif Config

func main() {
	// Here we define our flags, and bind them to viper configurations
	//pflag.StringVar(&cfgFile, "config", "", "config file")
	//
	//pflag.String("sys.loglevel", "default_flag_level", "loglevel")

	if _, err := viperx.BindAllFlags(pflag.CommandLine, Config{}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pflag.Parse()
	//
	if err := viperx.ParseConfig(&viperxConif, "DEMO", "demo/conf/viperx.json"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Config is: %#v\n", viperxConif)
	fmt.Printf("Config is: %#v\n", utils.ToString(viperxConif))

}
