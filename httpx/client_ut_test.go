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

	"github.com/madlabx/pkgx/log"
)

type ClientTestSuite struct {
	suite.Suite
}

func (ts *ClientTestSuite) SetupSuite() {
}

func (ts *ClientTestSuite) TearDownSuite() {
}

func (ts *ClientTestSuite) TestInvalidUrl() {
	require2 := require.New(ts.T())

	client, err := NewProxyClient("127.0.0.1", 1000000, true)
	require2.NotNil(err)
	stats, err := client.GetAndDrop("http://127.0.0.1:1")
	require2.NotNil(err)
	require2.Equal(stats.Error, ERR_CONNTION)
}

func (ts *ClientTestSuite) TestConnectFail() {
	require2 := require.New(ts.T())

	client, err := NewProxyClient("127.0.0.1", 1, true)
	require2.NotNil(err)
	stats, err := client.GetAndDrop("http://127.0.0.1:1")
	require2.NotNil(err)
	require2.Equal(stats.Error, ERR_CONNTION)
}

func (ts *ClientTestSuite) TestReadError() {
	t := require.New(ts.T())
	var proxy *httptest.Server

	server, proxy, client, err := createTestServerClient(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("server got one request %s", r.URL)
		_, _ = w.Write([]byte(fmt.Sprintf("hello from url: %s", r.URL)))
	}, func(w http.ResponseWriter, r *http.Request) {
		c := []byte(fmt.Sprintf("hello from url: %s", r.URL))
		repeat := 1000000
		size := len(c) * repeat
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		quit := make(chan struct{})
		go func() {
			for idx := 0; idx < repeat; idx++ {
				_, _ = w.Write(c)
			}
			quit <- struct{}{}
		}()

		time.Sleep(10 * time.Millisecond)
		proxy.CloseClientConnections()
		<-quit
	})

	t.Nil(err)

	defer server.Close()
	defer proxy.Close()

	log.Info("server url", server.URL)
	log.Info("proxy url", proxy.URL)

	stats, err := client.GetAndDrop(server.URL + "/test")
	t.NotNil(err)
	t.Equal(stats.Error, ERR_READ)

	time.Sleep(1 * time.Second)
}

func (ts *ClientTestSuite) TestGet() {
	t := require.New(ts.T())

	server, proxy, client, err := createTestServerClient(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("server got one request %s", r.URL)
		_, _ = w.Write([]byte(fmt.Sprintf("hello from url: %s", r.URL)))
	}, nil)

	t.Nil(err)

	defer server.Close()
	defer proxy.Close()

	log.Info("server url", server.URL)
	log.Info("proxy url", proxy.URL)

	stats, err := client.GetAndDrop(server.URL + "/test")
	t.Nil(err)
	t.Equal(stats.Error, ERR_NONE)

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
			if rasps, err := proxyClient.Do(proxyReq); err == nil {
				defer rasps.Body.Close()

				for key, vals := range rasps.Header {
					w.Header().Set(key, strings.Join(vals, ";"))
				}
				w.WriteHeader(rasps.StatusCode)
				_, _ = io.Copy(w, rasps.Body)
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
