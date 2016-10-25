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

//http://product-v3.soubarato.com.br/product?q=itemId:(122256872)&limit=1&paymentOptionIds=CARTAO_VISA,CARTAO_SUBA_MASTERCARD,BOLETO
type ProdutoB2w struct {
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

//image cnova http://www.pontofrio-imagens.com.br/a/1/499169539/9195122.jpg
//preco cnova http://preco.api-pontofrio.com.br/V1/Skus/PrecoVenda/?idssku=9195122
//detalhes cnova http://rec.pontofrio.com.br/productdetails/api/skusdetails/getbyids?ids=9195122
type ProdutoCNova struct {
	Valores []struct {
		PrecoVenda struct {
			Preco            float32 `json:"preco"`
			PontosMultiplos  int
			PrecoDe          float32
			PrecoTabela      float32
			PrecoSemDesconto float32
		}
	} `json:"PrecoSkus"`
	Detalhes []struct {
		NomeProduto      string `json:"ProductName"`
		Categoria        string `json:"CategoryName"`
		SubCategoria     string `json:"SubCategory"`
		SubCategoriaNome string `json:"SkuCategoryName"`
		IDImagemPadrao   string `json:"ImageFileId"`
		IDImagem130      string `json:"ImageFile130x130Id"`
		IDImagem45       string `json:"ImageFile45x45Id"`
		IDImagem90       string `json:"ImageFile90x90Id"`
		IDImage292       string `json:"ImageFile130x130Id"`
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
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Language", "pt-BR,pt;q=0.8,en-US;q=0.6,en;q=0.4,es;q=0.2")
		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		target := ProdutoB2w{}

		json.NewDecoder(res.Body).Decode(&target)
		fmt.Println(target)
	}

}
