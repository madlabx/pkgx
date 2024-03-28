package typex

import (
	"encoding/json"
	"strconv"
)

type Error string

func NewError(err error) Error {
	return Error(err.Error())
}

func (v Error) Error() string {
	return string(v)
}

func (v Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(v))
}

func (v *Error) UnmarshalJSON(data []byte) error {

	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*v = Error(*x)
	return nil
}

type JsonMap map[string]any

func NewJsonMap(m map[string]string) JsonMap {

	jm := JsonMap{}
	for k, v := range m {
		jm[k] = v
	}
	return jm
}

func (m JsonMap) Update(om JsonMap) {

	for k, v := range om {
		m[k] = v
	}
}

func (m JsonMap) GetString(name string) string {
	if v, ok := m[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (m JsonMap) GetStrings(name string) []string {
	if v, ok := m[name]; ok {
		if s, ok := v.([]string); ok {
			return s
		}
	}
	return nil
}

func (m JsonMap) GetBool(name string) bool {
	if v, ok := m[name]; ok {
		if s, ok := v.(bool); ok {
			return s
		}
	}
	return false
}

func (m JsonMap) GetInt(name string) int {
	return int(m.GetInt64(name))
}

func (m JsonMap) GetUint32(name string) uint32 {
	return uint32(m.GetInt64(name))
}

func (m JsonMap) GetInt64(name string) int64 {
	if v, ok := m[name]; ok {
		switch cv := v.(type) {
		case float32:
			return int64(cv)
		case float64:
			return int64(cv)
		case int64:
			return cv
		case int:
			return int64(cv)
		case string:
			i, err := strconv.ParseInt(cv, 10, 64)
			if err == nil {
				return i
			} else {
			}
		default:
		}
	}
	return 0
}

func (m JsonMap) GetObject(name string) JsonMap {
	if v, ok := m[name]; ok {
		if s, ok := v.(map[string]interface{}); ok {
			return JsonMap(s)
		}
	}

	return nil
}
