package main

import (
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/utils"
	"github.com/madlabx/pkgx/viperx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

type Config struct {
	Ttt string `vx_name:"" vx_short:"t" vx_default:"1234" vx_desc:"test for ttt"`
	Nc  struct {
		Sys   int64  `vx_flag:";;1001;bandwith"`
		Limit string `vx_name:"sys" vx_short:"l" vx_default:"100" vx_desc:"test for limit"`
	}
	Sys struct {
		LogLevel string
		Size     int
	}
}

var viperxConfig Config
var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "viperx",
	Short: "This is a demo",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Printf("Config is: %v", utils.ToString(viperxConfig))
	},
}

func main() {

	// 设置日志级别
	logrus.SetLevel(logrus.InfoLevel)

	// 创建一个TextFormatter并自定义输出格式
	// 例如，我们添加了时间戳和日志级别
	formatter := &log.TextFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		FullTimestamp:    true,
		ForceColors:      true,
		DisableColors:    false,
		DisableFileLine:  false,
		QuoteEmptyFields: true,
		DisableSorting:   true,
	}

	// 设置日志记录器的格式化器
	logrus.SetFormatter(formatter)

	// 输出日志到stdout
	logrus.SetOutput(os.Stdout)

	cobra.OnInitialize(initConfig)

	// Here we define our flags, and bind them to viper configurations
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.myapp.yaml)")
	//viper.BindPFlag("foo", rootCmd.PersistentFlags().Lookup("foo"))

	rootCmd.PersistentFlags().String("sys.loglevel", "flag_sys_dev", "A help for foo")

	rootCmd.Flags().StringP("keyPath", "s", "xx", "desc")

	//Bind command flags
	if _, err := viperx.BindAllFlags(rootCmd.Flags(), Config{}); err != nil {
		logrus.Println(err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		logrus.Println(err)
		os.Exit(1)
	}
}

// initConfig reads ENV variable and config file if set.
func initConfig() {
	if err := viperx.ParseConfig(&viperxConfig, "DEMO", "conf/viperx.json"); err != nil {
		logrus.Println(err)
	}
}
