package httpx

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/madlabx/pkgx/errors"

	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/typex"

	resty "github.com/go-resty/resty/v2"
)

var ErrBadParams error = errors.New("Bad input params")

// var ErrNot200OK error = errors.New("the response is not 200 OK")
var UserAgent string = "Json Http Client"
var StatsTotalReqs int64

type StatusError struct {
	Err    string
	Status int
	Rsp    *resty.Response
}

func NewStatusError(method, host, api string, rsp *resty.Response) *StatusError {

	if len(api) > 256 {
		api = api[:256]
	}
	s := fmt.Sprintf("%s %s %s response not ok", method, host, api)
	return &StatusError{
		Err:    s,
		Status: rsp.StatusCode(),
		Rsp:    rsp,
	}
}

func NewStatusError2(method, url string, rsp *resty.Response) *StatusError {

	if len(url) > 256 {
		url = url[:256]
	}
	s := fmt.Sprintf("%s %s response not ok", method, url)
	return &StatusError{
		Err:    s,
		Status: rsp.StatusCode(),
		Rsp:    rsp,
	}
}

func (e *StatusError) Error() string {

	if e.Status > 0 {
		body := e.Body()
		if len(body) > 256 {
			body = body[:256]
		}
		return e.Err + ", response code: " + strconv.Itoa(e.Status) + ", body: " + body
	}
	return e.Err
}

func (e *StatusError) Body() string {

	if e.Rsp != nil {
		return string(e.Rsp.Body())
	}

	return ""
}

// format := "${time_rfc3339} ${time_unix} ${status} ${method} ${latency} ${remote_ip} ${bytes_in} ${bytes_out} ${uri}\n"
type RequestStats struct {
	Scheme   string
	Host     string
	Port     int
	Uri      string
	Method   string
	ReqTime  time.Time
	SendTime time.Time
	RspTime  time.Time
	Latency  int64
	PreTime  int64
	Status   int
}

func (s *RequestStats) LogLine() string {
	str := s.ReqTime.Format("2006/01/02-15:04:05.000") + " " +
		strconv.Itoa(s.Status) + " " + s.Method + " " + strconv.FormatInt(s.Latency, 10) +
		" " + s.Host + " " + strconv.Itoa(s.Port) + " " + s.Uri + "\n"
	return str
}

type JsonClient struct {
	IsHttps bool
	Host    string
	Port    int
	Timeout int // in milliseconds

	statsChan chan<- *RequestStats
	c         *resty.Client
}

func NewJsonClient(host string, port int, timeout int) *JsonClient {

	return NewJsonClientWithRetry(host, port, timeout, 0, 0)
}

func NewJsonClientWithRetry(host string, port, timeout, retryCount, retryWaitTime int) *JsonClient {

	c := &JsonClient{
		Host:    host,
		Port:    port,
		Timeout: timeout,
		c:       resty.New(),
	}
	if c.Timeout > 0 {
		c.c.SetTimeout(time.Duration(c.Timeout) * time.Millisecond)
	}
	c.c.RetryCount = retryCount
	c.c.RetryWaitTime = time.Millisecond * time.Duration(retryWaitTime)
	c.c.AllowGetMethodPayload = true
	c.c.SetCloseConnection(true)
	return c
}

func (c *JsonClient) SetStatsChan(ch chan<- *RequestStats) {
	c.statsChan = ch
}

func (c *JsonClient) SetReuseConnection() {

	log.Infof("Set client %s:%d reuse connection", c.Host, c.Port)
	c.c.SetCloseConnection(false)
}

func (c *JsonClient) SetInsecure() {

	log.Infof("Set client %s:%d insecure", c.Host, c.Port)
	c.c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}

func (c *JsonClient) SetTransport(maxIdleConns, maxIdleConnsPerHost int) {

	log.Infof("Set client %s:%d transport, maxIdleConns: %d, maxIdleConnsPerHost: %d",
		c.Host, c.Port, maxIdleConns, maxIdleConnsPerHost)
	tspt := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(c.Timeout) * time.Millisecond,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   time.Duration(c.Timeout) * time.Millisecond,
		ExpectContinueTimeout: 1 * time.Second,
	}
	c.c.SetTransport(tspt)
}

func (c *JsonClient) Transport() http.RoundTripper {

	tspt := c.c.GetClient().Transport
	if tspt == nil {
		return http.DefaultTransport
	}
	return tspt
}

func (c *JsonClient) Client() *resty.Client {
	return c.c
}

func (c *JsonClient) Get(url string, headers map[string]string) (typex.JsonMap, error) {
	return c.Request("GET", url, headers, nil)
}

func (c *JsonClient) GetR(url string, headers map[string]string) (*resty.Response, error) {
	return c.RequestR(typex.JsonMap{}, "GET", url, headers, nil)
}

func (c *JsonClient) Post(url string, headers map[string]string, data interface{}) (typex.JsonMap, error) {
	return c.Request("POST", url, headers, data)
}

func (c *JsonClient) PostR(url string, headers map[string]string, data interface{}) (*resty.Response, error) {
	return c.RequestR(nil, "POST", url, headers, data)
}

func (c *JsonClient) Put(url string, headers map[string]string, data interface{}) (typex.JsonMap, error) {
	return c.Request("PUT", url, headers, data)
}

func (c *JsonClient) PutR(url string, headers map[string]string, data interface{}) (*resty.Response, error) {
	return c.RequestR(nil, "PUT", url, headers, data)
}

func (c *JsonClient) Del(url string, headers map[string]string, data interface{}) (typex.JsonMap, error) {
	return c.Request("DELETE", url, headers, data)
}

func (c *JsonClient) DelR(url string, headers map[string]string) (*resty.Response, error) {
	return c.RequestR(typex.JsonMap{}, "DELETE", url, headers, nil)
}

func (c *JsonClient) Request(method, url string, headers map[string]string,
	data interface{}) (typex.JsonMap, error) {

	return c.RequestTimeout(method, url, headers, data, -1)
}

func (c *JsonClient) StatusError(method, url string, rsp *resty.Response) *StatusError {

	return NewStatusError2(method, c.Url(url), rsp)
}

func (c *JsonClient) RequestTimeout(method, url string, headers map[string]string,
	data interface{}, timeout int) (typex.JsonMap, error) {

	rsp, err := c.requestRTimeout(typex.JsonMap{}, method, url, headers, data, timeout)
	if err != nil {
		return nil, err
	}

	ret := rsp.Result()
	code := rsp.StatusCode()
	if code != 200 && code != 201 {
		if ret != nil {
			return *ret.(*typex.JsonMap), c.StatusError(method, url, rsp)
		}
		return nil, c.StatusError(method, url, rsp)
	}

	return *ret.(*typex.JsonMap), nil
}

func (c *JsonClient) RequestR(result interface{}, method, url string, headers map[string]string,
	data interface{}) (*resty.Response, error) {

	return c.requestRTimeout(result, method, url, headers, data, -1)
}

func (c *JsonClient) requestRTimeout(result interface{}, method, url string, headers map[string]string,
	data interface{}, timeout int) (*resty.Response, error) {

	hc := c.c
	stats := &RequestStats{
		Scheme:  "http",
		Host:    c.Host,
		Port:    c.Port,
		Uri:     url,
		Method:  method,
		ReqTime: time.Now(),
	}
	if timeout > 0 {
		if timeout != c.Timeout {
			// different timeout, use new http client
			hc = resty.New()
			hc.SetTransport(c.Transport())
			hc.SetTimeout(time.Duration(timeout) * time.Millisecond)
		}
	} else {
		timeout = c.Timeout
	}
	var r *resty.Request
	if hc != nil {
		r = hc.R()
	} else {
		r = c.c.R()
	}
	if result != nil {
		r.SetResult(result)
	}
	if len(headers) == 0 {
		headers = map[string]string{
			"User-Agent": UserAgent,
		}
	} else {
		_, ok := headers["User-Agent"]
		if !ok {
			headers["User-Agent"] = UserAgent
		}
	}
	r.SetHeaders(headers)
	if data != nil {
		if headers["Content-Type"] == "application/x-www-form-urlencoded" {
			r.SetFormData(data.(map[string]string))
		} else {
			r.SetBody(data)
		}
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if len(c.Host) == 0 || c.Port == 0 {
			err := errors.New(fmt.Sprintf("The url: '%s' is invalid without host set in the client", url))
			log.Error(err)
			return nil, ErrBadParams
		}
		url = c.Url(url)
	}
	log.Debugf("Send api request: %s %s, timeout: %d", method, url, timeout)
	if c.IsHttps {
		stats.Scheme = "https"
	}
	defer func() {
		stats.PreTime = stats.SendTime.Sub(stats.ReqTime).Nanoseconds() / 1000
		stats.Latency = stats.RspTime.Sub(stats.ReqTime).Nanoseconds() / 1000
		if c.statsChan != nil {
			c.statsChan <- stats
		}
	}()
	atomic.AddInt64(&StatsTotalReqs, 1)
	rsp, err := r.Execute(method, url)
	stats.SendTime = r.Time
	stats.RspTime = time.Now()
	if err != nil {
		log.Errorf("%v, timeout: %d ms", err, timeout)
		return nil, err
	}
	log.Debugf("Recv api response: %s %s, status: %d", method, url, rsp.StatusCode())
	defer rsp.RawBody().Close()

	stats.Status = rsp.StatusCode()
	return rsp, nil
}

func (c *JsonClient) Url(api string) string {

	if strings.HasPrefix(api, "http://") || strings.HasPrefix(api, "https://") {
		return api
	}
	p := "http:"
	host := c.Host
	if c.IsHttps {
		p = "https:"
		if c.Port != 443 {
			host = fmt.Sprintf("%s:%d", host, c.Port)
		}
	} else {
		if c.Port != 80 {
			host = fmt.Sprintf("%s:%d", host, c.Port)
		}
	}
	return fmt.Sprintf("%s//%s%s", p, host, api)
}

type AuthJsonClient struct {
	Client JsonClient
	token  string
}

func (c *AuthJsonClient) Login(method, url string, headers map[string]string,
	data interface{}, tokenField string) (string, error) {

	rspData, err := c.Request(method, url, headers, data)
	if err != nil {
		return "", err
	}
	c.token = rspData.GetString(tokenField)
	if len(c.token) == 0 {
		err = errors.New("No such token field: " + tokenField)
		log.Error(err)
		return "", err
	}
	return c.token, nil
}

func (c *AuthJsonClient) Get(url string, headers map[string]string) (typex.JsonMap, error) {
	return c.Request("GET", url, headers, nil)
}

func (c *AuthJsonClient) Post(url string, headers map[string]string, data interface{}) (typex.JsonMap, error) {
	return c.Request("POST", url, headers, data)
}

func (c *AuthJsonClient) Put(url string, headers map[string]string, data interface{}) (typex.JsonMap, error) {
	return c.Request("PUT", url, headers, data)
}

func (c *AuthJsonClient) DelR(url string, headers map[string]string) (*resty.Response, error) {
	return c.RequestR(typex.JsonMap{}, "DELETE", url, headers, nil)
}

func (c *AuthJsonClient) Request(method, url string, headers map[string]string,
	data interface{}) (typex.JsonMap, error) {

	rsp, err := c.RequestR(typex.JsonMap{}, method, url, headers, data)
	if err != nil {
		return nil, err
	}
	return *rsp.Result().(*typex.JsonMap), nil
}

func (c *AuthJsonClient) RequestR(result interface{}, method, url string, headers map[string]string,
	data interface{}) (*resty.Response, error) {

	if len(c.token) > 0 {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Authorization"] = "Bearer " + c.token
	}
	return c.Client.RequestR(result, method, url, headers, data)
}

func IsConnError(err error) bool {

	if ue, ok := err.(*url.Error); ok {
		err = ue.Err
	}
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		//println("Timeout")
		return true
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			//println("Unknown host")
			return true
		} else if t.Op == "read" {
			//println("Connection refused")
			return true
		}
	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			//println("Connection refused")
			return true
		}
	}

	return false
}
