package main

import (
	"fmt"
	"github.com/madlabx/pkgx/viperx"
	"github.com/spf13/cobra"
	"os"
)

type ConfigSys struct {
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

var viperxConfig Config
var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "viperx",
	Short: "This is a demo",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Config is: %#v\n", viperxConfig)
	},
}

func main() {
	cobra.OnInitialize(initConfig)

	// Here we define our flags, and bind them to viper configurations
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myapp.yaml)")
	//viper.BindPFlag("foo", rootCmd.PersistentFlags().Lookup("foo"))

	rootCmd.PersistentFlags().String("sys.loglevel", "flag_sys_dev", "A help for foo")

	rootCmd.Flags().StringP("keyPath", "s", "xx", "desc")

	//Bind command flags
	if _, err := viperx.BindAllFlags(rootCmd.Flags(), Config{}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads ENV variable and config file if set.
func initConfig() {
	if err := viperx.ParseConfig(&viperxConfig, "DEMO", "../conf/viperx.json"); err != nil {
		fmt.Sprintln(err)
	}
}
