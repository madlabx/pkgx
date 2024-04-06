package viperx

import (
	"reflect"
	"strings"

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
	tagViperXFieldShort        = "vx_short"
	//tagViperXFieldRange        = "vx_range"
)

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

func parseFlatStyle(vxTag string) (string, string, string, string) {
	kv := strings.Split(vxTag, ";")
	if len(kv) == 4 {
		return kv[0], kv[1], kv[2], kv[3]
	} else {
		return "", "", "", ""
	}
}

func parseFlagOpts(tag reflect.StructTag) (string, string, string, string) {
	vxTagString, ok := tag.Lookup(tagViperX)
	if ok {
		//Address  string `vx_tag:"address;a;127.0.0.1;address to listen on"`
		return parseFlatStyle(vxTagString)
	} else {
		//Address  string `vx_name:"address" vx_short:"a" vx_default:"127.0.0.1" vx_desc:"address to listen on"`
		return tag.Get(tagViperXFieldName), tag.Get(tagViperXFieldShort), tag.Get(tagViperXFieldDefault), tag.Get(tagViperXFieldDescription)
	}
}

// TODO 是否可以用DecodeHookFunc更优雅地实现?
func parse(fs *pflag.FlagSet, rt reflect.Type, tagName string, parts ...string) error {
	var (
		err error
	)
	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i)
		fieldName := parseTypeName(t, tagName)

		switch t.Type.Kind() {
		case reflect.Struct: // Handle nested struct
			if len(fieldName) == 0 {
				err = parse(fs, t.Type, tagName, parts...)
			} else {
				err = parse(fs, t.Type, tagName, append(parts, fieldName)...)
			}
		default: // Handle leaf field
			keyPath := strings.Join(append(parts, fieldName), ".")
			vxName, vxShort, vxDefault, vxDesc := parseFlagOpts(t.Tag)

			if len(vxName) == 0 {
				fs.StringP(keyPath, vxShort, vxDefault, vxDesc)
				if err = vx.v.BindPFlag(keyPath, fs.Lookup(keyPath)); err != nil {
					return err
				}
				vx.v.SetDefault(keyPath, vxDefault)
			} else {
				fs.StringP(vxName, vxShort, vxDefault, vxDesc)
				if err = vx.v.BindPFlag(keyPath, fs.Lookup(vxName)); err != nil {
					return err
				}
				vx.v.SetDefault(keyPath, vxDefault)
			}
		}
	}

	return err
}
