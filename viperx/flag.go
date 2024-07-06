package viperx

import (
	"reflect"
	"strings"

	"github.com/madlabx/pkgx/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// TODO 增加must，必须项
var (
	defaultMapStructureTagName = "mapstructure"
	tagViperX                  = "vx_tag"
	tagViperXFieldName         = "vx_name"
	tagViperXFieldDefault      = "vx_default"
	tagViperXFieldDescription  = "vx_desc"
	tagViperXFieldMust         = "vx_must"
	tagViperXFieldShort        = "vx_short"
	tagViperXFieldRange        = "vx_range"
)

const VxFlagsRangeSeparator = ","

type vxFlags struct {
	Path    string
	Name    string `json:",omitempty"`
	Short   string `json:",omitempty"`
	Default string `json:",omitempty"`
	Desc    string `json:",omitempty"`
	Must    string `json:",omitempty"`
	Range   string `json:",omitempty"`
}

func (vx *vxFlags) SetPFlag(fs *pflag.FlagSet) {
	fs.StringP(vx.Name, vx.Short, vx.Default, vx.Desc)
}

func (vx *vxFlags) String() string {
	return utils.ToString(vx)
}

func getMapStructureTagName(opts ...viper.DecoderConfigOption) string {
	c := &mapstructure.DecoderConfig{}
	for _, opt := range opts {
		opt(c)
	}

	if c.TagName == "" {
		return defaultMapStructureTagName
	}
	return c.TagName
}

func parseTypeName(t reflect.StructField, tagName string) string {
	tv, ok := t.Tag.Lookup(tagName)
	if !ok {
		return t.Name
	} else if tv == ",squash" || tv == "" {
		return ""
	} else {
		return tv
	}
}

func parseFlatStyle(vxTag string) vxFlags {
	kv := strings.Split(vxTag, ";")
	if len(kv) == 6 {
		//TODO prompt some error
		return vxFlags{Name: kv[0], Short: kv[1], Default: kv[2], Desc: kv[3], Must: kv[4], Range: kv[5]}
	} else {
		return vxFlags{}
	}
}

func parseFlagOpts(tag reflect.StructTag) vxFlags {
	vxTagString, ok := tag.Lookup(tagViperX)
	if ok {
		//Address  string `vx_tag:"address;a;127.0.0.1;address to listen on"`
		return parseFlatStyle(vxTagString)
	} else {
		//Address  string `vx_name:"address" vx_short:"a" vx_default:"127.0.0.1" vx_desc:"address to listen on"`
		return vxFlags{
			Name:    tag.Get(tagViperXFieldName),
			Short:   tag.Get(tagViperXFieldShort),
			Default: tag.Get(tagViperXFieldDefault),
			Desc:    tag.Get(tagViperXFieldDescription),
			Must:    tag.Get(tagViperXFieldMust),
			Range:   tag.Get(tagViperXFieldRange)}
	}
}

// TODO 是否可以用DecodeHookFunc更优雅地实现?
func (o *ViperX) parse(fs *pflag.FlagSet, rt reflect.Type, tagName string, parts ...string) error {
	var (
		err error
	)
	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i)
		fieldName := parseTypeName(t, tagName)

		switch t.Type.Kind() {
		case reflect.Struct: // Handle nested struct
			if len(fieldName) == 0 {
				err = o.parse(fs, t.Type, tagName, parts...)
			} else {
				err = o.parse(fs, t.Type, tagName, append(parts, fieldName)...)
			}
		default: // Handle leaf field
			keyPath := strings.Join(append(parts, fieldName), ".")
			flags := parseFlagOpts(t.Tag)
			flags.Path = keyPath
			if flags.Name == "" {
				flags.Name = keyPath
			}

			flags.SetPFlag(fs)
			if err = vx.v.BindPFlag(keyPath, fs.Lookup(flags.Name)); err != nil {
				return err
			}
			vx.v.SetDefault(keyPath, flags.Default)

			if flags.Must == "true" && flags.Default == "" {
				o.mustList = append(o.mustList, &flags)
			}

			if flags.Range != "" {
				o.rangeList = append(o.rangeList, &flags)
			}
		}
	}

	return err
}
