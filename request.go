package sopromocao

import(
    "net/http"
    "net"
    "time"
    "encoding/json"
)

//proxyUrl, _ := url.Parse("tcp://192.168.56.1:8888")
var netTransport = &http.Transport{
	//Proxy: http.ProxyURL(proxyUrl),
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}
var client = &http.Client{
	Timeout:   time.Second * 5,
	Transport: netTransport,
}


func Request(url string, target interface{}) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.8,en-US;q=0.6,en;q=0.4,es;q=0.2")
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&target)
}