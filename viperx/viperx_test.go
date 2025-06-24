package viperx

import (
	"log"
	"testing"

	"github.com/spf13/viper"
)

func TestTomls(t *testing.T) {
	viper.SetConfigFile("")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Can't read config:%v", err)
	}

	log.Printf("sys.logdir:%v", GetString("sys.logdir", "./"))
}
