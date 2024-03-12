package auth

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/madlabx/pkgx/viperx"
)

type SignCfg struct {
	SignEnable bool     `json:"signEnable,omitempty"`
	SignFormat []string `json:"signFormat,omitempty"`
	SignSecret string   `json:"signSecret,omitempty"`
	SignAlgo   string   `json:"signAlgo,omitempty"`
	SignEnc    string   `json:"signAlgo,omitempty"`
	SignExpire int      `json:"signExpire,omitempty"`
}

func (sc *SignCfg) Merge(sc2 *SignCfg) {

	if !sc2.SignEnable {
		sc.SignEnable = false
	}
	if len(sc.SignFormat) == 0 {
		sc.SignFormat = sc2.SignFormat
	}
	if sc.SignSecret == "" {
		sc.SignSecret = sc2.SignSecret
	}
	if sc.SignAlgo == "" {
		sc.SignAlgo = sc2.SignAlgo
	}
	if sc.SignEnc == "" {
		sc.SignEnc = sc2.SignEnc
	}
	if sc.SignExpire == 0 {
		sc.SignExpire = sc2.SignExpire
	}
}

func (sc *SignCfg) Validate() error {

	if !sc.SignEnable {
		return nil
	}
	if sc.SignSecret == "" {
		return fmt.Errorf("The signSecret must be set")
	}
	return nil
}

func ParseSignCfg(prefix string) *SignCfg {

	cfg := &SignCfg{
		SignEnable: viper.GetBool(prefix + ".signEnable"),
		SignFormat: ParseSignFormat(viperx.GetString(prefix+".signFormat", "$path$expires$algo")),
		SignSecret: viper.GetString(prefix + ".signSecret"),
		SignAlgo:   viperx.GetString(prefix+".signAlgo", "hmac-sha256"),
		SignEnc:    viperx.GetString(prefix+".signEnc", "hex"),
		SignExpire: viperx.GetInt(prefix+".signExpire", 30),
	}
	return cfg
}
