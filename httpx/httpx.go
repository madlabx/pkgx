package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/madlabx/pkgx/errno"
	"github.com/madlabx/pkgx/log"
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

var defaultClient = &Client{cli: http.DefaultClient}

func ResetDefaultClient(hc *Client) {
	defaultClient = hc
}

func HttpGetBody(url string) (*http.Response, []byte, error) {
	return requestBytesForBody(defaultClient, "GET", url, nil, true)
}

func HttpGet(url string) (*http.Response, error) {
	return requestBytes(defaultClient, "GET", url, nil)
}

func HttpPostBody(url string, body interface{}) (*http.Response, []byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Parse json failed, url: %s, obj: %#v", url, body)
		return nil, nil, err
	}

	return requestBytesForBody(defaultClient, "POST", url, b, true)
}

type Client struct {
	cli *http.Client
}

func (hc *Client) clone() *Client {
	newRawClient := *hc.cli
	return &Client{cli: &newRawClient}
}

func (hc *Client) HttpPost(url string, body interface{}) (*http.Response, error) {
	return httpPostInternal(hc, url, body)
}

func (hc *Client) HttpGetBody(url string) (*http.Response, []byte, error) {
	return requestBytesForBody(hc, "GET", url, nil, true)
}

func (hc *Client) HttpGet(url string) (*http.Response, error) {
	return requestBytes(hc, "GET", url, nil)
}

func (hc *Client) WithTimeout(timeout int) *Client {
	hc.cli.Timeout = time.Second * time.Duration(timeout)
	return hc
}

func WithTimeout(timeout int) *Client {
	return defaultClient.clone().WithTimeout(timeout)
}

func HttpPost(url string, body interface{}) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		log.Errorf("Parse json failed, url: %s, obj: %#v", url, body)
		//return nil, ErrorResp(http.StatusBadRequest, errno.ECODE_BAD_REQUEST_PARAM, err)
		return nil, err
	}

	return requestBytes(
		defaultClient,
		"POST",
		url,
		b)
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

	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Errorf("read body err, err:%v, response:%v", err.Error(), rsp)
			return nil, nil, err
		}

		err = errors.New(rsp.Status)

		return rsp, body, err
	}

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
		return nil, nil, wrap(err)
	}
	defer func() {
		if rsp != nil {
			_ = rsp.Body.Close()
		}
	}()

	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			log.Errorf("read body err, err:%v, response:%v", err.Error(), rsp)
			return nil, nil, errors.New("Failed to parse response body")
		}

		//log.Errorf("Got not 200 response[%#v], body[%v]", rsp, string(body))

		var newStatusError JsonResponse
		err = json.Unmarshal(body, &newStatusError)
		if err != nil {
			log.Errorf("read body err, err:%v, response:%v", err.Error(), rsp)
			return nil, nil, errors.New("Failed to parse error information: " + string(body))
		}
		newStatusError.Status = rsp.StatusCode

		if newStatusError.Errno == errno.ECODE_SUCCESS {
			newStatusError.Errno = errno.ECODE_FAILED_HTTP_REQUEST
			newStatusError.WithMsg(string(body))
		}
		//log.Errorf("err:%v", newStatusError)

		return rsp, body, &newStatusError
	}

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
