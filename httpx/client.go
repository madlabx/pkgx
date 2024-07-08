package httpx

//import (

//)
//
//func NewProxyTransport(proxyHost string, proxyPort int, isHttp bool) (*http.Transport, error) {
//	var url string
//
//	if isHttp {
//		url = fmt.Sprintf("http://%s:%d", proxyHost, proxyPort)
//	} else {
//		url = fmt.Sprintf("https://%s:%d", proxyHost, proxyPort)
//	}
//
//	proxyUrl, cause := urllib.Parse(url)
//	if cause != nil {
//		return nil, cause
//	}
//
//	return &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, nil
//}
//
//func NewProxyClient(proxyHost string, proxyPort int, isHttp bool) (*HttpClient, error) {
//	trspt, cause := NewProxyTransport(proxyHost, proxyPort, isHttp)
//	if cause != nil {
//		return nil, cause
//	}
//
//	return NewHttpClient(trspt), nil
//}
//
//type HttpClient struct {
//	client *http.Client
//}
//
//func NewHttpClient(transport http.RoundTripper) *HttpClient {
//	return &HttpClient{
//		client: &http.Client{Transport: transport, Timeout: 3 * time.Second},
//	}
//}
//
//func (pc *HttpClient) Get(url string, out io.Writer) (*Stats, error) {
//	stats := NewStats()
//	stats.Url = url
//
//	lastTime := time.Now()
//	req, cause := http.NewRequest("GET", url, nil)
//	if cause != nil {
//		stats.Error = ERR_CONNTION
//		return stats, cause
//	}
//	req.Header.Set("User-Agent", "Content Preposition Client")
//
//	resp, cause := pc.client.Do(req)
//	if cause != nil {
//		stats.Error = ERR_REQ
//		return stats, cause
//	}
//
//	defer resp.Body.Close()
//
//	n, cause := io.Copy(out, resp.Body)
//	if cause != nil {
//		stats.Error = ERR_READ
//		return stats, cause
//	}
//
//	stats.DownloadSize = n
//
//	pc.fillRespStats(stats, resp, lastTime)
//
//	return stats, nil
//}
//
//func (pc *HttpClient) PostJson(url string, obj interface{}) (*Stats, error) {
//	stats := NewStats()
//	stats.Url = url
//
//	lastTime := time.Now()
//	b, cause := json.Marshal(obj)
//	if cause != nil {
//		stats.Error = ERR_JSON
//		return nil, cause
//	}
//
//	resp, cause := pc.client.Post(url, "application/json", bytes.NewBuffer(b))
//	if cause != nil {
//		stats.Error = ERR_REQ
//		return nil, cause
//	}
//
//	pc.fillRespStats(stats, resp, lastTime)
//	return stats, nil
//}
//
//func (pc *HttpClient) GetAndDrop(url string) (*Stats, error) {
//	out := &DropWriter{}
//
//	return pc.Get(url, out)
//}
//
//func (pc *HttpClient) Purge(url string) (*Stats, error) {
//	stats := NewStats()
//	stats.Url = url
//
//	lastTime := time.Now()
//	req, cause := http.NewRequest("PURGE", url, nil)
//	if cause != nil {
//		stats.Error = ERR_CONNTION
//		return stats, cause
//	}
//	req.Header.Set("User-Agent", "Content Preposition Client")
//
//	resp, cause := pc.client.Do(req)
//	if cause != nil {
//		stats.Error = ERR_CONNTION
//		return stats, cause
//	}
//
//	defer resp.Body.Close()
//
//	pc.fillRespStats(stats, resp, lastTime)
//
//	return stats, nil
//}
//
//func (pc *HttpClient) fillRespStats(stats *Stats, resp *http.Response, reqTime time.Time) {
//	stats.ContentLength = resp.ContentLength
//	stats.Status = resp.StatusCode
//	stats.Proto = resp.Proto
//	stats.RespHeader = resp.Header
//
//	spent := time.Since(reqTime)
//	stats.TimeToServe = spent.Seconds()
//	stats.Resp = resp
//}
