package viperx

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// Unmarshal decodes the configuration into a struct using viper.Unmarshal.
// It accepts any type of rawVal where configuration data will be stored, and opts for decoder options.
func Unmarshal(rawVal any, opts ...viper.DecoderConfigOption) error {
	return viper.Unmarshal(rawVal, opts...)
}

// Init initializes the configuration by setting up flags, environment variables, and configuration files.
// It takes a FlagSet for command-line arguments, a prefix for environment variables, and a path to the configuration file.
func Init(cmdFlags *pflag.FlagSet, envPrefix string, cfgFile string) error {
	if err := InitFlags(cmdFlags); err != nil {
		return err
	}

	InitEnvs(envPrefix, ".", "_")

	if err := InitConfig(cfgFile, ",", "", ""); err != nil {
		return err
	}

	return nil
}

// InitFlags binds a set of command-line flags to their corresponding configuration keys in viper.
// If the flag has not been changed and has a non-zero length default value, it will mark the flag as changed.
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

// InitEnvs sets up environment variables to override configuration values.
// It takes a prefix for environment variable names, a delimiter for configuration keys, and an environment delimiter for replacing in keys.
func InitEnvs(prefix, keyDelimiter, envDelimiter string) {
	viper.AutomaticEnv() // automatically override values with those from the environment
	viper.SetEnvPrefix(prefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(keyDelimiter, envDelimiter))

	prefix = prefix + "_"

	// Traverse all environment variables
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]

		// Check if the environment variable has the required prefix
		if len(pair) > 1 && len(pair[1]) > 0 && strings.HasPrefix(key, prefix) {
			// Bind the current environment variable
			envKey := key[len(prefix):]                                         // Remove the prefix
			configKey := strings.ReplaceAll(envKey, envDelimiter, keyDelimiter) // Replace the delimiter, commonly changing '_' to '.'
			_ = viper.BindEnv(configKey)

			// Optionally set a default value to ensure it appears in AllKeys()
			// viper.SetDefault(configKey, "")
		}
	}
}

// InitConfig initializes configuration files using viper.
// It takes paths to configuration file, file name, and file type.
// If a configuration file is found, it will be read into viper.
func InitConfig(cfgFile, cfgFilePath, cfgFileName, cfgFileType string) error {
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
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func ConfigFileUsed() string {
	return viper.ConfigFileUsed()
}

// SetConfigFile sets the path to the configuration file.
func SetConfigFile(in string) {
	viper.SetConfigFile(in)
}

// AddConfigPath adds a new path for viper to search for the configuration file in.
func AddConfigPath(in string) {
	viper.AddConfigPath(in)
}

// SetConfigName sets the name for the configuration file.
func SetConfigName(in string) {
	viper.SetConfigName(in)
}

// SetConfigType sets the type of the configuration file.
func SetConfigType(in string) {
	viper.SetConfigType(in)
}

// GetString retrieves a string value from the configuration.
// It returns a default value if the key is not set.
func GetString(name string, def string) string {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetString(name)
}

// GetStrings retrieves a slice of strings from the configuration.
// It returns a default value if the key is not set.
func GetStrings(name string, def []string) []string {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetStringSlice(name)
}

// GetInt retrieves an integer value from the configuration.
// It returns a default value if the key is not set.
func GetInt(name string, def int) int {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetInt(name)
}

// GetInt64 retrieves an int64 value from the configuration.
// It returns a default value if the key is not set.
func GetInt64(name string, def int64) int64 {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetInt64(name)
}

// GetBool retrieves a boolean value from the configuration.
// It returns a default value if the key is not set.
func GetBool(name string, def bool) bool {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetBool(name)
}

// GetFloat64 retrieves a float64 value from the configuration.
// It returns a default value if the key is not set.
func GetFloat64(name string, def float64) float64 {
	if !viper.IsSet(name) {
		return def
	}
	return viper.GetFloat64(name)
}
