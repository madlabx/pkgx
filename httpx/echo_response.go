package httpx

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
)

func NewEtag(modTime time.Time, length int64) string {
	timestampHex := strconv.FormatInt(modTime.Unix(), 16)
	// 将长度转换为16进制
	lengthHex := strconv.FormatInt(length, 16)
	// 将两部分用'-'连接
	return "\"" + timestampHex + "-" + lengthHex + "\""
}

// CheckIfNoneMatch if Etag same, true
func CheckIfNoneMatch(r *http.Request, currentEtag string) bool {
	inm := r.Header.Get("If-None-Match")
	if inm == "" {
		return false
	}
	return etagWeakMatch(inm, currentEtag)
}

// etagWeakMatch reports whether a and b match using weak ETag comparison.
// Assumes a and b are valid ETags.
func etagWeakMatch(a, b string) bool {
	return strings.TrimPrefix(a, "W/") == strings.TrimPrefix(b, "W/")
}

var unixEpochTime = time.Unix(0, 0)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

// CheckIfModifiedSince if not modified, return true
func CheckIfModifiedSince(r *http.Request, modtime time.Time) bool {
	if r.Method != "GET" && r.Method != "HEAD" {
		return false
	}
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" || isZeroTime(modtime) {
		return false
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return false
	}
	// The Last-Modified header truncates sub-second precision so
	// the modTime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if ret := modtime.Compare(t); ret <= 0 {
		return true
	}
	return false
}

func SendResp(c echo.Context, resp error) (err error) {
	if c.Response().Committed {
		return resp
	}

	if resp == nil {
		rid := errCodeDic.NewRequestId()
		c.Response().Header().Set(echo.HeaderXRequestID, rid)
		return c.NoContent(http.StatusOK)
	}

	jr := Wrap(resp)
	if jr.RequestId == "" {
		jr.RequestId = errCodeDic.NewRequestId()
	}
	c.Response().Header().Set(echo.HeaderXRequestID, jr.RequestId)

	return jr.cjson(c)
}

func ServeContent(w http.ResponseWriter, req *http.Request, name string, modTime time.Time, length int64, content io.ReadSeeker) {
	rid := errCodeDic.NewRequestId()
	w.Header().Set(echo.HeaderXRequestID, rid)
	w.Header().Set("Etag", NewEtag(modTime, length))

	http.ServeContent(w, req, name, modTime, content)
}

func ServeContentWithTag(w http.ResponseWriter, req *http.Request, name string, modTime time.Time, localEtag string, content io.ReadSeeker) {
	rid := errCodeDic.NewRequestId()
	w.Header().Set(echo.HeaderXRequestID, rid)
	w.Header().Set("Etag", localEtag)
	http.ServeContent(w, req, name, modTime, content)
}
