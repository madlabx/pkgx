package viperx

import "github.com/mitchellh/mapstructure"

var EnableReportErrorUnused = func(c *mapstructure.DecoderConfig) {
	c.ErrorUnused = true
}
