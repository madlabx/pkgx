package httpx

import (
	"errors"
	resty "github.com/go-resty/resty/v2"
	"github.com/madlabx/pkgx/auth"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/typex"
	"math/rand"
	"net"
	"sync"
	"time"
)

type RrServer struct {
	host     string
	deadTime int64
	offline  bool
}

type RoundrobinClient struct {
	service   string
	retryTime int
	sign      *auth.SignCfg

	lock    sync.RWMutex
	servers []*RrServer
	clients []*JsonClient
	curIdx  int
	client  *JsonClient
}

func NewRoundrobinClient(service string, https bool, hosts []string, port int,
	sign *auth.SignCfg, apiTimeout, retryTime int) *RoundrobinClient {

	client := &RoundrobinClient{
		service:   service,
		retryTime: retryTime,
		sign:      sign,
		servers:   []*RrServer{},
		client:    nil,
	}
	for _, host := range hosts {
		server := &RrServer{
			host: host,
		}
		client.servers = append(client.servers, server)
		c := NewJsonClient(host, port, apiTimeout)
		c.IsHttps = https
		client.clients = append(client.clients, c)
	}
	client.next()

	return client
}

func NewRoundrobinClient2(service string, cl []*JsonClient, sign *auth.SignCfg, retryTime int) *RoundrobinClient {

	client := &RoundrobinClient{
		service:   service,
		retryTime: retryTime,
		sign:      sign,
		clients:   cl,
		client:    nil,
	}
	for _, jc := range cl {
		client.servers = append(client.servers, &RrServer{
			host: jc.Host,
		})
	}
	client.next()

	return client
}

// must called before any request
func (c *RoundrobinClient) SetReuseConnection() {

	for _, jc := range c.clients {
		jc.SetReuseConnection()
	}
}

func (c *RoundrobinClient) SetInsecure() {

	for _, jc := range c.clients {
		jc.SetInsecure()
	}
}

// must called before any request
func (c *RoundrobinClient) SetTransport(maxIdleConns, maxIdleConnsPerHost int) {

	for _, jc := range c.clients {
		jc.SetTransport(maxIdleConns, maxIdleConnsPerHost)
	}
}

func (c *RoundrobinClient) StatusError(method, api string, rsp *resty.Response) *StatusError {

	client := c.Client()
	if client != nil {
		return client.StatusError(method, api, rsp)
	}

	return NewStatusError(method, "", api, rsp)
}

func (c *RoundrobinClient) Client() *JsonClient {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.client
}

func (c *RoundrobinClient) next() *JsonClient {
	c.lock.Lock()
	defer c.lock.Unlock()

	var cfg *RrServer
	for cnt := 0; cnt < len(c.servers); cnt += 1 {
		if c.curIdx >= len(c.servers) {
			c.curIdx = 0
		}
		server := c.servers[c.curIdx]
		if server.offline {
			now := time.Now().Unix()
			if now-server.deadTime >= int64(c.retryTime) {
				server.offline = false
				c.curIdx += 1
				cfg = server
				break
			}
			c.curIdx += 1
			continue
		}

		cfg = server
		c.client = c.clients[c.curIdx]
		c.curIdx += 1
		break
	}

	if cfg == nil {
		log.Warnf("No online server available, try to use one dead server")
		c.curIdx = rand.Intn(len(c.servers))
		c.client = c.clients[c.curIdx]
	}
	log.Infof("Now use %s host: %s, port: %d", c.service, c.client.Host, c.client.Port)
	return c.client
}

func (c *RoundrobinClient) Get(url string, headers map[string]string) (typex.JsonMap, error) {
	return c.Request("GET", url, headers, nil)
}

func (c *RoundrobinClient) GetR(url string, headers map[string]string) (*resty.Response, error) {
	return c.RequestR(typex.JsonMap{}, "GET", url, headers, nil)
}

func (c *RoundrobinClient) Post(url string, headers map[string]string, data interface{}) (typex.JsonMap, error) {
	return c.Request("POST", url, headers, data)
}

func (c *RoundrobinClient) PostR(url string, headers map[string]string, data interface{}) (*resty.Response, error) {
	return c.RequestR(nil, "POST", url, headers, data)
}

func (c *RoundrobinClient) Put(url string, headers map[string]string, data interface{}) (typex.JsonMap, error) {
	return c.Request("PUT", url, headers, data)
}

func (c *RoundrobinClient) PutR(url string, headers map[string]string, data interface{}) (*resty.Response, error) {
	return c.RequestR(nil, "PUT", url, headers, data)
}

func (c *RoundrobinClient) DelR(url string, headers map[string]string) (*resty.Response, error) {
	return c.RequestR(typex.JsonMap{}, "DELETE", url, headers, nil)
}

func (c *RoundrobinClient) Request(method, url string, headers map[string]string,
	data interface{}) (typex.JsonMap, error) {

	rsp, err := c.RequestR(typex.JsonMap{}, method, url, headers, data)
	if err != nil {
		return nil, err
	}

	code := rsp.StatusCode()
	if code != 200 && code != 201 {
		host := ""
		client := c.Client()
		if client != nil {
			host = client.Host
		}
		return nil, NewStatusError(method, host, url, rsp)
	}

	return *rsp.Result().(*typex.JsonMap), nil
}

// light wrapper of RoundrobinClient.RequestR
func (c *RoundrobinClient) RequestR(result interface{}, method, url string, headers map[string]string, data interface{}) (*resty.Response, error) {
	client := c.Client()
	if client == nil {
		client = c.next()
	}

	// round robin try next server
	cnt := 0
	for client != nil {
		cnt += 1
		if cnt >= 3 || cnt > len(c.servers) {
			break
		}
		api, err := c.urlsign(client, url)
		if err != nil {
			log.Errorf("Sign url %s failed, %v", url, err)
			return nil, err
		}
		rsp, err := client.RequestR(result, method, api, headers, data)
		// TODO: only retry on specific error?
		if _, ok := err.(net.Error); ok {
			log.Errorf("Retry %s due to error: %s", c.service, err)
			c.markDead(client.Host)
			client = c.next()
			continue
		}
		return rsp, err
	}

	return nil, errors.New("No available servers for " + c.service)
}

func (c *RoundrobinClient) Url(api string) string {

	js := c.Client()
	if js != nil {
		return js.Url(api)
	}
	return c.clients[0].Url(api)
}

func (c *RoundrobinClient) urlsign(jc *JsonClient, api string) (string, error) {

	sign := c.sign
	if sign == nil || !sign.SignEnable {
		return api, nil
	}
	url, err := auth.UrlSign(jc.Url(api), map[string]string{}, sign.SignFormat,
		sign.SignAlgo, sign.SignEnc, sign.SignSecret, sign.SignExpire)

	if err != nil {
		log.Errorf("Sign api %s failed: %v", api, err)
	}
	return url, err
}

func (c *RoundrobinClient) markDead(host string) {

	for _, server := range c.servers {
		if server.host == host {
			server.offline = true
			server.deadTime = time.Now().Unix()
			log.Infof("Mark the %s server %s dead", c.service, host)
		}
	}
}
