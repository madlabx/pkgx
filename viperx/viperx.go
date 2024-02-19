package viperx

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func Unmarshal(rawVal any, opts ...viper.DecoderConfigOption) error {
	return viper.Unmarshal(rawVal, opts...)
}

func Init(cmdFlags *pflag.FlagSet, envPrefix string, cfgFile string) error {
	if err := InitFlags(cmdFlags); err != nil {
		return err
	}

	InitEnvs(envPrefix, ".", "_")
	InitConfig(cfgFile, ",", "", "")

	return nil
}

func InitFlags(cmdFlags *pflag.FlagSet) error {
	//Make sure the default value also make sense
	cmdFlags.VisitAll(func(f *pflag.Flag) {
		if !f.Changed && len(f.Value.String()) != 0 {
			f.Changed = true
		}
	})
	if cmdFlags != nil {
		return viper.BindPFlags(cmdFlags)
	}
	return nil
}

func InitEnvs(prefix, keyDelimiter, envDelimiter string) {
	viper.AutomaticEnv() // automatically override values with those from the environment
	viper.SetEnvPrefix(prefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(keyDelimiter, envDelimiter))

	prefix = prefix + "_"

	// 遍历所有环境变量
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]

		// 检查环境变量是否具有所需的前缀
		if len(pair) > 1 && len(pair[1]) > 0 && strings.HasPrefix(key, prefix) {
			// 绑定当前环境变量
			envKey := key[len(prefix):]                                         // 移除前缀
			configKey := strings.ReplaceAll(envKey, envDelimiter, keyDelimiter) //替换分割符号，一般是把_换为.
			_ = viper.BindEnv(configKey)

			// 可以选择设置一个默认值，确保它出现在 AllKeys() 中
			//viper.SetDefault(configKey, "")
		}
	}
}

func InitConfig(cfgFile, cfgFilePath, cfgFileName, cfgFileType string) {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	} else {
		if cfgFilePath != "" {
			viper.AddConfigPath(cfgFilePath)
		}
		if cfgFileName != "" { // adding home directory as first search path
			viper.SetConfigName(cfgFileName)
		}
		if cfgFileType != "" {
			viper.SetConfigType(cfgFileType)
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// SetConfigFile Inherit from viper
func SetConfigFile(in string) {
	viper.SetConfigFile(in)
}

func AddConfigPath(in string) {
	viper.AddConfigPath(in)
}

func SetConfigName(in string) {
	viper.SetConfigName(in)
}
func SetConfigType(in string) {
	viper.SetConfigType(in)
}

// GetString Expand func with default value
func GetString(name string, def string) string {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetString(name)
}

func GetStrings(name string, def []string) []string {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetStringSlice(name)
}

func GetInt(name string, def int) int {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetInt(name)
}

func GetInt64(name string, def int64) int64 {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetInt64(name)
}

func GetBool(name string, def bool) bool {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetBool(name)
}

func GetFloat64(name string, def float64) float64 {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetFloat64(name)
}
