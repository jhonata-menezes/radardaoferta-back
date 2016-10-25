package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type ProdutosB2w struct {
	Products []struct {
		ID      string `json:"id"`
		Nome    string `json:"name"`
		Imagens []struct {
			Medium     string
			Big        string
			Large      string
			ExtraLarge string
		} `json:"images"`
		Offers []struct {
			PaymentOptions struct {
				CartaoSubmarino struct {
					Price float32
				} `json:"CARTAO_SUBA_MASTERCARD"`
				CartaoVisa struct {
					Price float32
				} `json:"CARTAO_VISA"`
				Boleto struct {
					Price float32
				} `json:"BOLETO"`
			}
		}
	}
}

func main() {
	inicio := time.Now()
	var wg sync.WaitGroup
	urls := make(chan string, 100)
	wg.Add(1)
	go request(urls, &wg)
	for i := 0; i < 100; i++ {
		urls <- os.Args[1]
	}
	close(urls)
	wg.Wait()
	fmt.Println(time.Since(inicio))
}

func request(urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
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

	for url := range urls {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "pt-BR,pt;q=0.8,en-US;q=0.6,en;q=0.4,es;q=0.2")
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		target := ProdutosB2w{}
		//http://product-v3.soubarato.com.br/product?q=itemId:(122256872)&limit=1&paymentOptionIds=CARTAO_VISA,CARTAO_SUBA_MASTERCARD,BOLETO
		json.NewDecoder(res.Body).Decode(&target)
		fmt.Println(target)
	}

}
