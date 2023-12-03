package httpx

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	urllib "net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"risk_manager/pkg/log"
)

type ClientTestSuite struct {
	suite.Suite
}

func (ts *ClientTestSuite) SetupSuite() {
}

func (ts *ClientTestSuite) TearDownSuite() {
}

func (ts *ClientTestSuite) TestInvalidUrl() {
	require := require.New(ts.T())

	client, err := NewProxyClient("127.0.0.1", 1000000, true)

	stats, err := client.GetAndDrop("http://127.0.0.1:1")
	require.NotNil(err)
	require.Equal(stats.Error, ERR_CONNTION)
}

func (ts *ClientTestSuite) TestConnectFail() {
	require := require.New(ts.T())

	client, err := NewProxyClient("127.0.0.1", 1, true)

	stats, err := client.GetAndDrop("http://127.0.0.1:1")
	require.NotNil(err)
	require.Equal(stats.Error, ERR_CONNTION)
}

func (ts *ClientTestSuite) TestReadError() {
	require := require.New(ts.T())
	var proxy *httptest.Server

	server, proxy, client, err := createTestServerClient(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("server got one request %s", r.URL)
		w.Write([]byte(fmt.Sprintf("hello from url: %s", r.URL)))
	}, func(w http.ResponseWriter, r *http.Request) {
		c := []byte(fmt.Sprintf("hello from url: %s", r.URL))
		repeat := 1000000
		size := len(c) * repeat
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		quit := make(chan struct{})
		go func() {
			for idx := 0; idx < repeat; idx++ {
				w.Write(c)
			}
			quit <- struct{}{}
		}()

		time.Sleep(10 * time.Millisecond)
		proxy.CloseClientConnections()
		<-quit
	})

	require.Nil(err)

	defer server.Close()
	defer proxy.Close()

	log.Info("server url", server.URL)
	log.Info("proxy url", proxy.URL)

	stats, err := client.GetAndDrop(server.URL + "/test")
	require.NotNil(err)
	require.Equal(stats.Error, ERR_READ)

	time.Sleep(1 * time.Second)
}

func (ts *ClientTestSuite) TestGet() {
	require := require.New(ts.T())

	server, proxy, client, err := createTestServerClient(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("server got one request %s", r.URL)
		w.Write([]byte(fmt.Sprintf("hello from url: %s", r.URL)))
	}, nil)

	require.Nil(err)

	defer server.Close()
	defer proxy.Close()

	log.Info("server url", server.URL)
	log.Info("proxy url", proxy.URL)

	stats, err := client.GetAndDrop(server.URL + "/test")
	require.Nil(err)
	require.Equal(stats.Error, ERR_NONE)

	time.Sleep(1 * time.Second)
}

func createTestServerClient(serverHandler func(http.ResponseWriter, *http.Request),
	proxyHandler func(http.ResponseWriter, *http.Request)) (*httptest.Server, *httptest.Server, *HttpClient, error) {

	server := httptest.NewServer(http.HandlerFunc(serverHandler))

	if proxyHandler == nil {
		proxyHandler = func(w http.ResponseWriter, r *http.Request) {
			url := r.URL.String()
			proxyClient := &http.Client{}
			proxyReq, err := http.NewRequest(r.Method, url, r.Body)
			if err != nil {
				log.Error("NewRequest failed with error", err)
				return
			}

			proxyReq.Header = r.Header
			if rsps, err := proxyClient.Do(proxyReq); err == nil {
				defer rsps.Body.Close()

				for key, vals := range rsps.Header {
					w.Header().Set(key, strings.Join(vals, ";"))
				}
				w.WriteHeader(rsps.StatusCode)
				io.Copy(w, rsps.Body)
			} else {
				log.Error("proxy failed to get url", url, "with error", err)
			}
		}
	}

	proxy := httptest.NewServer(http.HandlerFunc(proxyHandler))

	pu, _ := urllib.Parse(proxy.URL)
	segs := strings.Split(pu.Host, ":")
	var port int64 = 80
	if len(segs) > 1 {
		port, _ = strconv.ParseInt(segs[1], 10, 32)
	}

	client, err := NewProxyClient(segs[0], int(port), true)

	if err != nil {
		server.Close()
		proxy.Close()
		return nil, nil, nil, err
	}

	return server, proxy, client, nil
}

func TestClientTestSuite_UT(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
