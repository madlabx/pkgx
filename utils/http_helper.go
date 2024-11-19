package utils

import (
	"net/http"
	"strings"
)

func TrimHttpStatusText(status int) string {
	trimmedSpace := strings.Replace(http.StatusText(status), " ", "", -1)
	trimmedSpace = strings.Replace(trimmedSpace, "-", "", -1)
	return trimmedSpace
}
