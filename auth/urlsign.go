package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"math/rand"
	URL "net/url"
	"strconv"
	"strings"
	"time"

	"risk_manager/pkg/log"
)

func ValidateSign(url string, format []string, enc, secret string) error {

	purl, err := URL.Parse(url)
	if err != nil {
		return fmt.Errorf("Invalid url: %s, error: %v", url, err)
	}
	idx := strings.Index(url, "?")
	if idx >= 0 {
		url = url[:idx]
	}
	// fill the sign fields
	fields := map[string]string{
		"url":  url,
		"algo": "hmac-sha256",
	}
	fillSignFieldsFromUrl(purl, fields)
	if es, ok := fields["expires"]; ok {
		expires, err := strconv.ParseInt(es, 10, 64)
		if err != nil {
			log.Debugf("Invalid expires value: %d, error: %v", es, err)
			return fmt.Errorf("Invalid url expires param: %s, error: %v", es, err)
		}
		now := time.Now().UTC().Unix()
		if expires <= now {
			log.Debugf("The url %s expires, expires value: %d, now: %d", url, expires, now)
			return fmt.Errorf("Url expires")
		}
	}

	usign, _ := fields["sign"]
	if usign == "" {
		return fmt.Errorf("No sign found in url")
	}

	sign, err := genSign(fields, format, fields["algo"], enc, secret)
	if err != nil {
		return err
	}

	if usign != sign {
		log.Debugf("The url sign %s mismatches with gen sign: %s", usign, sign)
		return fmt.Errorf("The url sign is invalid")
	}

	return nil
}

func UrlSign(url string, fields map[string]string, format []string,
	algo, enc, secret string, expire int) (string, error) {

	purl, err := URL.Parse(url)
	if err != nil {
		return "", fmt.Errorf("Invalid url: %s, error: %v", url, err)
	}
	idx := strings.Index(url, "?")
	ourl := url
	if idx >= 0 {
		url = url[:idx]
	}
	// fill the sign fields
	now := time.Now().UTC().Unix()
	fields["url"] = url
	fields["algo"] = algo
	fields["expires"] = fmt.Sprintf("%d", now+int64(expire))
	fm := map[string]string{}
	fillSignFieldsFromUrl(purl, fm)
	for k, v := range fields {
		fm[k] = v
	}
	fields = fm

	sign, err := genSign(fields, format, algo, enc, secret)
	if err != nil {
		return "", err
	}

	querys := []string{}
	for _, field := range format {
		if field == "url" || field == "uri" || field == "path" {
			continue
		}
		if _, ok := purl.Query()[field]; ok {
			// skip the field already in the url
			continue
		}

		value := fields[field]
		querys = append(querys, field+"="+value)
	}
	querys = append(querys, "sign="+sign)
	queryStr := strings.Join(querys, "&")
	if idx >= 0 {
		return ourl + "&" + queryStr, nil
	}
	return ourl + "?" + queryStr, nil
}

func fillSignFieldsFromUrl(purl *URL.URL, fields map[string]string) {

	fields["uri"] = purl.Path
	fields["path"] = purl.Path
	fields["host"] = purl.Hostname()
	fields["port"] = purl.Port()
	fields["scheme"] = purl.Scheme
	fields["query"] = purl.RawQuery

	for k, vl := range purl.Query() {
		if len(vl) == 0 {
			continue
		}
		fields[k] = vl[0]
	}
}

func genSign(fields map[string]string, format []string, algo, enc, secret string) (string, error) {

	signs := []string{}
	for _, field := range format {
		v, ok := fields[field]
		if !ok {
			log.Debugf("No such field '%s' value provided for urlsign, use the empty string instead", field)
		}
		signs = append(signs, v)
	}
	ss := strings.Join(signs, "/")
	signStr := ss + " " + secret
	var h hash.Hash
	switch algo {
	case "md5":
		h = md5.New()
	case "hmac-md5":
		signStr = ss
		h = hmac.New(md5.New, []byte(secret))
	case "sha1":
		h = sha1.New()
	case "hmac-sha1":
		signStr = ss
		h = hmac.New(sha1.New, []byte(secret))
	case "sha256":
		h = sha256.New()
	case "hmac-sha256":
		signStr = ss
		h = hmac.New(sha256.New, []byte(secret))
	case "sha512":
		h = sha512.New()
	case "hmac-sha512":
		signStr = ss
		h = hmac.New(sha512.New, []byte(secret))
	default:
		log.Errorf("Unsupported urlsign algo: %s", algo)
		return "", fmt.Errorf("Unsupported urlsign algo: %s", algo)
	}

	h.Write([]byte(signStr))
	bytes := h.Sum(nil)
	sign := EncSign(bytes, enc)

	log.Debugf("sign string(%s/%s): %s, sign: %s, signs: %#v, format: %#v",
		algo, enc, ss, sign, signs, format)

	return sign, nil
}

func EncSign(bytes []byte, enc string) string {

	switch enc {
	case "base64":
		bs := base64.StdEncoding.EncodeToString(bytes)
		bs = strings.Replace(bs, "+", "-", -1)
		bs = strings.Replace(bs, "/", "_", -1)
		bs = strings.Replace(bs, "=", "", -1)
		return bs
	default:
		// use the hex as default
		return fmt.Sprintf("%x", bytes)
	}
}

var letters = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func genTxid(now int64, deviceId string) string {
	return fmt.Sprintf("%s-%d-%s", deviceId, now, RandString(10))
}

func ParseSignFormat(fstr string) []string {

	format := strings.Split(fstr, "$")
	es := ""
	as := ""
	nfmt := []string{}
	for _, fmt := range format {
		if fmt == "expires" {
			es = fmt
		} else if fmt == "algo" {
			as = fmt
		}
		if fmt != "" {
			nfmt = append(nfmt, fmt)
		}
	}
	format = nfmt
	if as == "" {
		format = append(format, "algo")
	}
	if es == "" {
		format = append(format, "expires")
	}

	return format
}
