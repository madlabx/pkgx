package httpx

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
)

type Client struct {
	cli *http.Client
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.cli.Do(req)
}

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

var defaultClient = &Client{cli: http.DefaultClient}

func ResetDefaultClient(hc *Client) {
	defaultClient = hc
}

func (hc *Client) WithTimeout(timeout int) *Client {
	hc.cli.Timeout = time.Second * time.Duration(timeout)
	return hc
}

func NewClientWithTimeout(timeout int) *Client {
	return defaultClient.clone().WithTimeout(timeout)
}

func (hc *Client) HttpPostBody(url string, body interface{}) (*http.Response, []byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Parse json failed, url: %s, obj: %#v", url, body)
		return nil, nil, err
	}

	return requestBytesForBody(hc, "POST", url, b, true)
}

func PostX(url string, reqBody interface{}, result interface{}) (*JsonResponse, error) {
	return defaultClient.PostX(url, reqBody, result)
}

func (hc *Client) PostX(url string, reqBody interface{}, result interface{}) (*JsonResponse, error) {
	b, err := json.Marshal(reqBody)
	if err != nil {
		log.Errorf("Parse json failed, url: %s, obj: %#v", url, reqBody)
		return nil, errors.Wrap(err)
	}

	return hc.PostBytesX(url, b, result)
}

func PostBytesX(url string, b []byte, result interface{}) (*JsonResponse, error) {
	return defaultClient.PostBytesX(url, b, result)
}
func (hc *Client) PostBytesX(url string, b []byte, result interface{}) (*JsonResponse, error) {
	resp, body, err := requestBytesForBody(hc, "POST", url, b, true)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	jr := &JsonResponse{}
	jr.Result = result
	err = json.Unmarshal(body, &jr)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	jr.Status = resp.StatusCode
	if !jr.IsOK() {
		jr.Result = nil
	}

	return jr, nil
}

func (hc *Client) clone() *Client {
	newRawClient := *hc.cli
	return &Client{cli: &newRawClient}
}

func HttpGetBody(url string) (*http.Response, []byte, error) {
	return defaultClient.HttpGetBody(url)
}

func (hc *Client) HttpGetBody(url string) (*http.Response, []byte, error) {
	return requestBytesForBody(hc, "GET", url, nil, true)
}

func HttpGet(url string) (*http.Response, error) {
	return defaultClient.HttpGet(url)
}

func (hc *Client) HttpGet(url string) (*http.Response, error) {
	return requestBytes(hc, "GET", url, nil)
}
func HttpPostBody(url string, body interface{}) (*http.Response, []byte, error) {
	return defaultClient.HttpPostBody(url, body)
}

func HttpPost(url string, body interface{}) (*http.Response, error) {
	return defaultClient.HttpPost(url, body)
}
func (hc *Client) HttpPost(url string, body interface{}) (*http.Response, error) {
	return httpPostInternal(hc, url, body)
}

func httpPostInternal(cli *Client, url string, body interface{}) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Parse json failed, url: %s, obj: %#v", url, body)
		//return nil, ErrorResp(http.StatusBadRequest, errno.ECODE_BAD_REQUEST_PARAM, err)
		return nil, err
	}

	return requestBytes(cli, "POST", url, b)
}

func requestBytes(cli *Client, method, url string, bodyBytes []byte) (*http.Response, error) {
	resp, _, err := requestBytesForBody(cli, method, url, bodyBytes, false)
	return resp, err
}

func requestBytesForBody(hc *Client, method, requrl string, bodyBytes []byte, wantBody bool) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, requrl, bytes.NewReader(bodyBytes))

	if err != nil {
		log.Errorf("failed to build request, err:%#v", err.Error())
		return nil, nil, err
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Connection", "close")
	rsp, err := hc.cli.Do(req)
	if err != nil {
		//TODO add httpError
		return nil, nil, err
	}
	defer func() {
		if rsp != nil {
			_ = rsp.Body.Close()
		}
	}()

	if wantBody {
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Errorf("read body err, %v", err.Error())
			return nil, nil, err
		}
		return rsp, body, err
	}

	return rsp, nil, err
}

func requestBytesForBodyNormal(method, reqUrl string, bodyBytes []byte, wantBody bool) (*http.Response, []byte, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest(method, reqUrl, bytes.NewReader(bodyBytes))

	if err != nil {
		log.Errorf("failed to build request, err:%#v", err.Error())
		return nil, nil, err
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Connection", "close")
	rsp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to send request, err:%#v", err.Error())
		return nil, nil, Wrap(err)
	}
	defer func() {
		if rsp != nil {
			_ = rsp.Body.Close()
		}
	}()

	if wantBody {
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Errorf("read body err, %v", err.Error())
			//return nil, nil, errStrResp(rsp.StatusCode, errno.ECODE_FAILED_HTTP_REQUEST, "Failed to parse response body")
			return nil, nil, errors.New("Failed to parse response body")
		}
		return rsp, body, err
	}

	return rsp, nil, err
}

func ResponseToMap(body []byte) (map[string]interface{}, error) {
	var set map[string]interface{}
	if err := json.Unmarshal(body, &set); err != nil {
		log.Errorf("Unmarshal err, %v", err.Error())
		return nil, err
	}
	return set, nil
}
