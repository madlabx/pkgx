package httpx

import (
	"net/http"
)

const (
	ERR_NONE     = "none"
	ERR_CONNTION = "Connection Error"
	ERR_READ     = "Read Content Error"
	ERR_REQ      = "Request Error"
	ERR_JSON     = "Encode/Decode json Error"
)

type Stats struct {
	Error         string
	Url           string
	Status        int
	Proto         string
	ContentLength int64
	DownloadSize  int64
	TimeToServe   float64 // in seconds
	RespHeader    http.Header
	Resp          *http.Response
}

func NewStats() *Stats {
	return &Stats{
		Error: ERR_NONE,
	}
}
