package viperx

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ViperX struct {
	v         *viper.Viper
	mustList  []*vxFlags
	rangeList []*vxFlags
	mutex     sync.Mutex
	//flags *pflag.FlagSet
}

var (
	vx *ViperX
)

func init() {
	vx = New()
}

func GetViper() *viper.Viper {
	return vx.v
}

func New() *ViperX {
	return &ViperX{
		v: viper.GetViper(),
	}
}

func (o *ViperX) BindFlags(fs *pflag.FlagSet) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	//Make sure the default value in flag also make sense
	fs.VisitAll(func(f *pflag.Flag) {
		if !f.Changed && len(f.Value.String()) != 0 {
			o.v.SetDefault(f.Name, f.DefValue)
		}
	})

	return nil
}

func (o *ViperX) BindEnvs(prefix, keyDelimiter, envDelimiter string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.v.AutomaticEnv() // automatically override values with those from the environment
	o.v.SetEnvPrefix(prefix)
	o.v.SetEnvKeyReplacer(strings.NewReplacer(keyDelimiter, envDelimiter))

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
			_ = o.v.BindEnv(configKey)

			// Optionally set a default value to ensure it appears in AllKeys()
			// o.v.SetDefault(configKey, "")
		}
	}
}

func (o *ViperX) Parse(fs *pflag.FlagSet, cfg any, opts ...viper.DecoderConfigOption) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.parse(fs, reflect.TypeOf(cfg), getMapStructureTagName(opts...))
}

func doValidate(vld *validator.Validate, in any, rangeRequired string) (msg string, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprintf("invalid vx_range '%s', should be 'oneof=apple banana' or 'gte=5,lte=10'", rangeRequired)
			err = fmt.Errorf("%v", r)
		}
	}()

	err = vld.Var(in, rangeRequired)
	if err != nil {
		msg = fmt.Sprintf("invalid value '%v', should follow vx_range '%v'", in, rangeRequired)
	} else {
		err = nil
	}

	return
}

func (o *ViperX) Validate() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	var (
		err           error
		msg           string
		foundMustErr  = false
		foundRangeErr = false
	)
	vld := validator.New()
	bf := bytes.NewBuffer(make([]byte, 256))

	for _, mustItem := range o.mustList {
		if reflect.ValueOf(viper.Get(mustItem.Path)).IsZero() {
			if !foundMustErr {
				bf.WriteString("configure items without value:\n")
				foundMustErr = true
			}
			bf.WriteString("  " + mustItem.String() + "\n")
		}
	}

	for _, rangeItem := range o.rangeList {
		v := viper.Get(rangeItem.Path)

		if msg, err = doValidate(vld, v, rangeItem.Range); err != nil {
			if !foundRangeErr {
				bf.WriteString("configure items with invalid value:\n")
				foundRangeErr = true
			}
			bf.WriteString(fmt.Sprintf("  {\"Path\":\"%s\", \"Message\":\"%s\", \"RawError\":\"%s\"}\n", rangeItem.Path, msg, err.Error()))
		}
	}

	//Do release
	o.mustList = nil
	o.rangeList = nil

	if foundMustErr || foundRangeErr {
		return errors.New(bf.String())
	}

	return nil
}

// BindEnvs sets up environment variables to override configuration values.
// It takes a prefix for environment variable names, a delimiter for configuration keys, and an environment delimiter for replacing in keys.
func BindEnvs(prefix, keyDelimiter, envDelimiter string) {
	vx.BindEnvs(prefix, keyDelimiter, envDelimiter)
}

// Unmarshal decodes the configuration into a struct using viper.Unmarshal.
// It accepts any type of rawVal where configuration data will be stored, and opts for decoder options.
func Unmarshal(cfg any, opts ...viper.DecoderConfigOption) (err error) {
	return vx.v.Unmarshal(cfg, opts...)
}

// BindAllFlags 添加cfg结构体中vx_flag标记的Flag，并返回完整的FlagSet
// （推荐）若未定义name，name解析为cfg结构体成员名，多级使用"."相连
// 否则，解析为name
// 若fs为空，会初始化一个新的
func BindAllFlags(fs *pflag.FlagSet, cfg any, opts ...viper.DecoderConfigOption) (*pflag.FlagSet, error) {
	if fs == nil {
		fs = pflag.NewFlagSet("viperx", pflag.ContinueOnError)
	} else if err := vx.BindFlags(fs); err != nil {
		return fs, err
	}

	if err := vx.Parse(fs, cfg, opts...); err != nil {
		return nil, err
	}

	return fs, nil
}

// ParseConfig parse config in order by cli flags, env, config, default
func ParseConfig(cfg any, envPrefix string, cfgFile string, opts ...viper.DecoderConfigOption) error {
	//TODO make sure BindAllFlags before
	BindEnvs(envPrefix, ".", "_")
	if err := InitConfigFile(cfgFile, ",", "", ""); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg, opts...); err != nil {
		return err
	}

	return vx.Validate()
}

// InitConfigFile initializes configuration files using viper.
// It takes paths to configuration file, file name, and file type.
// If a configuration file is found, it will be read into viper.
func InitConfigFile(cfgFile, cfgFilePath, cfgFileName, cfgFileType string) error {
	if cfgFile != "" { // enable ability to specify config file via flag
		vx.v.SetConfigFile(cfgFile)
	} else {
		if cfgFilePath != "" {
			vx.v.AddConfigPath(cfgFilePath)
		}
		if cfgFileName != "" { // adding home directory as first search path
			vx.v.SetConfigName(cfgFileName)
		}
		if cfgFileType != "" {
			vx.v.SetConfigType(cfgFileType)
		}
	}

	// If a config file is found, read it in.
	if err := vx.v.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func BindPFlag(key string, flag *pflag.Flag) error { return vx.v.BindPFlag(key, flag) }

func BindPFlags(flags *pflag.FlagSet) error { return vx.v.BindPFlags(flags) }

func ConfigFileUsed() string {
	return vx.v.ConfigFileUsed()
}

// SetConfigFile sets the path to the configuration file.
func SetConfigFile(in string) {
	vx.v.SetConfigFile(in)
}

// AddConfigPath adds a new path for viper to search for the configuration file in.
func AddConfigPath(in string) {
	vx.v.AddConfigPath(in)
}

// SetConfigName sets the name for the configuration file.
func SetConfigName(in string) {
	vx.v.SetConfigName(in)
}

// SetConfigType sets the type of the configuration file.
func SetConfigType(in string) {
	vx.v.SetConfigType(in)
}

func ReadInConfig() error {
	return vx.v.ReadInConfig()
}

// GetString retrieves a string value from the configuration.
// It returns a default value if the key is not set.
func GetString(name string, def string) string {
	rst := vx.v.GetString(name)
	if len(rst) == 0 {
		return def
	}

	return rst
}

// GetStrings retrieves a slice of strings from the configuration.
// It returns a default value if the key is not set.
func GetStrings(name string, def []string) []string {
	if !vx.v.IsSet(name) {
		return def
	}
	return vx.v.GetStringSlice(name)
}

// GetInt retrieves an integer value from the configuration.
// It returns a default value if the key is not set.
func GetInt(name string, def int) int {
	if !vx.v.IsSet(name) {
		return def
	}
	return vx.v.GetInt(name)
}

// GetInt64 retrieves an int64 value from the configuration.
// It returns a default value if the key is not set.
func GetInt64(name string, def int64) int64 {
	if !vx.v.IsSet(name) {
		return def
	}
	return vx.v.GetInt64(name)
}

// GetBool retrieves a boolean value from the configuration.
// It returns a default value if the key is not set.
func GetBool(name string, def bool) bool {
	if !vx.v.IsSet(name) {
		return def
	}
	return vx.v.GetBool(name)
}

// GetFloat64 retrieves a float64 value from the configuration.
// It returns a default value if the key is not set.
func GetFloat64(name string, def float64) float64 {
	if !vx.v.IsSet(name) {
		return def
	}
	return vx.v.GetFloat64(name)
}
