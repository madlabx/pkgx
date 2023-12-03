package httpx

import (
	"net"
	"net/http"
	"strings"
)

func GetRealIp(req *http.Request) string {

	hip := req.Header.Get("X-Forwarded-For")
	if hip != "" {
		idx := strings.Index(hip, ",")
		if idx > 0 {
			return hip[:idx]
		}
		return hip
	}

	hip = req.Header.Get("X-Real-IP")
	if hip != "" {
		return hip
	}

	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}
