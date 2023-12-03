package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
}

func TestUrlSign(t *testing.T) {
	require := require.New(t)
	fs := "$url$vhost$deviceType$deviceId$profile$vendor$algo$expires$txid"
	format := ParseSignFormat(fs)
	secret := "123456"
	algo := "md5"
	enc := "hex"
	expire := 10
	fields := map[string]string{
		"algo":       algo,
		"vhost":      "test.com",
		"deviceType": "",
		"deviceId":   "00001",
		"profile":    "HD",
		"vendor":     "test",
	}
	url := "http://127.0.0.1:8900/v1/test?field1=hello"
	for _, algo := range []string{"md5", "sha1", "sha256", "sha512", "hmac-md5", "hmac-sha1", "hmac-sha256", "hmac-sha512"} {
		nurl, err := UrlSign(url, fields, format, algo, enc, secret, expire)
		require.Nil(err)
		err = ValidateSign(nurl, format, enc, secret)
		require.Nil(err)
	}

	enc = "base64"
	nurl, err := UrlSign(url, fields, format, algo, enc, secret, expire)
	require.Nil(err)
	err = ValidateSign(nurl, format, enc, secret)
	require.Nil(err)
}
