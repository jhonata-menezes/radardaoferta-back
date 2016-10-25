package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	urls := make(chan string, 100)
	wg.Add(1)
	go request(urls, &wg)
	urls <- os.Args[1]
	close(urls)
	wg.Wait()
}

func request(urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	for url := range urls {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36")
		response, err := client.Do(req)
		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		html := string(bytes)
		//rTitle, _ := regexp.Compile("mp\\-tit\\-name prodTitle\" title=\"(.*?)\"")
		rPrice, _ := regexp.Compile("content=\"[0-9]+\\.[0-9]+\">R\\$(.*?)<")

		//title := rTitle.FindStringSubmatch(html)
		price := rPrice.FindStringSubmatch(html)

		fmt.Println(price)
	}

}
