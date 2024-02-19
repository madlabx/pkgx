package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"pkgx/viperx"
)

var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "This is my application which is awesome",
	Run: func(cmd *cobra.Command, args []string) {
		var config Config
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Printf("Unable to decode into struct, %v", err)
		}

		fmt.Printf("Config is: %#v\n", config)

	},
}

type ConfigSys struct {
	LogLevel string
	Size     int
}

type Config struct {
	Sys  ConfigSys
	Sys2 ConfigSys
	Sys3 ConfigSys
}

func main() {
	cobra.OnInitialize(initConfigSimple)

	// Here we define our flags, and bind them to viper configurations
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myapp.yaml)")
	//viper.BindPFlag("foo", rootCmd.PersistentFlags().Lookup("foo"))

	rootCmd.PersistentFlags().String("sys.loglevel", "flag_sys_dev", "A help for foo")
	//rootCmd.Flags().String("sys.loglevel", "flag_sys_dev", "A help for foo")
	rootCmd.PersistentFlags().String("sys2.loglevel", "flag_sys2_dev", "A help for foo")
	rootCmd.Flags().String("sys3.loglevel", "flag_default_sys3_dev", "A help for foo")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfigSimple() {
	if err := viperx.Init(rootCmd.Flags(), "DEMO", "conf/viperx_demo.json"); err != nil {
		fmt.Sprintln(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := viperx.InitFlags(rootCmd.Flags()); err != nil {
		fmt.Sprintln(err)
	}

	viperx.InitEnvs("DEMO", ".", "_")

	viperx.InitConfig("", "conf/", "viperx_demo", "")
}
