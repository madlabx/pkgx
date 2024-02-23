package viperx

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"reflect"
	"strings"
)

// TODO 增加must，必须项
var (
	DefaultMapStructureTagName = "mapstructure"
	TagViperX                  = "vxflag"
	FieldName                  = "name"
	FieldDefault               = "default"
	FieldDescription           = "desc"
	FieldShort                 = "short"
)

func getMapStructureTagName(opts ...viper.DecoderConfigOption) string {
	c := &mapstructure.DecoderConfig{}
	for _, opt := range opts {
		opt(c)
	}

	if c.TagName == "" {
		return DefaultMapStructureTagName
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
			vxTagString, ok := t.Tag.Lookup(TagViperX)
			if !ok {
				continue
			}

			keyPath := strings.Join(append(parts, fieldName), ".")

			vxFlagOpts := strings.Split(vxTagString, ";")
			vxFlagName := ""
			vxFlagDefault := ""
			vxFlagDesc := ""
			vxFlagShort := ""
			for _, opt := range vxFlagOpts {
				kv := strings.Split(opt, ":")
				if len(kv) == 2 {
					if kv[0] == FieldName {
						//TODO validate name
						vxFlagName = kv[1]
					} else if kv[0] == FieldDefault {
						vxFlagDefault = kv[1]
					} else if kv[0] == FieldDescription {
						vxFlagDesc = kv[1]
					} else if kv[0] == FieldShort {
						vxFlagShort = kv[1]
					}
				}
			}

			if len(vxFlagName) == 0 {
				fs.StringP(keyPath, vxFlagShort, vxFlagDefault, vxFlagDesc)
			} else {
				fs.StringP(vxFlagName, vxFlagShort, vxFlagDefault, vxFlagDesc)
				if err = vx.v.BindPFlag(keyPath, fs.Lookup(vxFlagName)); err != nil {
					return err
				}
			}
		}
	}

	return err
}
