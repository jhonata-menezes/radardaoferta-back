package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	netUrl "net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const cnovaPrefixImg = "http://www.casasbahia-imagens.com.br/a/1/"

//http://product-v3.soubarato.com.br/product?q=itemId:(122256872)&limit=1&paymentOptionIds=CARTAO_VISA,CARTAO_SUBA_MASTERCARD,BOLETO
type ProdutoB2w struct {
	Link     string
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
	Link    string
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
		IDImagemPadrao   int    `json:"ImageFileId"`
		IDImagem130      int    `json:"ImageFile130x130Id"`
		IDImagem45       int    `json:"ImageFile45x45Id"`
		IDImagem90       int    `json:"ImageFile90x90Id"`
		IDImagem292      int    `json:"ImageFile292x292Id"`
	}
}

type ProdutoGenerico struct {
	IDProduto string
	Nome      string
	Valor     float32
	Imagens   []string
	Link      string
	Loja      string
}

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

func main() {
	inicio := time.Now()
	var wg sync.WaitGroup
	urls := make(chan string, 100)
	wg.Add(1)
	go processador(urls, &wg)
	for i := 0; i < 1; i++ {
		urls <- os.Args[1]
		urls <- os.Args[2]
		urls <- os.Args[3]
	}
	close(urls)
	wg.Wait()
	fmt.Println(time.Since(inicio))
}

func processador(urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urls {
		p := ProdutoCNova{}
		p.Link = url
		u := CnovaUrlToApi(url)
		request(u[0], &p)
		request(u[1], &p.Detalhes)
		if len(p.Valores) >= 1 {
			//Produto := lojaCnovaParaGenerico(p)
			//fmt.Printf("%#v", Produto)
			//key: value fmt.Printf("%#v", Produto)
		} else {
			log.Println("URL informada nao existe", url)
		}
	}

}

func request(url string, target interface{}) {
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

func identifyNomeLoja(url string) string {
	urlLoja, err := netUrl.Parse(url)
	if err != nil {
		panic(err)
	}
	return urlLoja.Host
	switch urlLoja.Host {
	case "pontofrio.com.br":
		return urlLoja.Host
	case "extra.com.br":
		//
	case "casasbahia.com.br":
		//
	case "cdiscount.com.br":
		//
	case "submarino.com.br":
		//
	case "americanas.com.br":
		//
	case "shoptime.com.br":
		//
	case "soubarato.com.br":
		//
	default:
		//
	}
	return ""
}

func validUrl(url string) bool {
	_, err := netUrl.Parse(url)
	if err != nil {
		return false
	} else {
		return true
	}
}

func lojaCnovaImagemMount(id int, codProduto string) string {
	return (cnovaPrefixImg + strconv.Itoa(id) + "/" + codProduto + ".jpg")
}

// EX: http://www.casasbahia.com.br/UtilidadesDomesticas/Panelas/PanelasdePressao/Panela-de-Pressao-em-Aluminio-Polido-45L-com-Visor-6014-MTA-4645916.html?recsource=whome&rectype=w17
// EX: http://www.pontofrio.com.br/TelefoneseCelulares/AcessoriosparaCelulares/acessorioscelularesSamsung/Fone-De-Ouvido-Samsung-Earpods-Volume-Microfone-9441277.html?recsource=whome&rectype=w17
func identifyCodProdutoCnova(url string) string {
	a, err := regexp.Compile("\\-([0-9]+)\\.html")
	if err != nil {
		panic(err)
	}
	result := a.FindStringSubmatch(url)
	if len(result) >= 2 {
		return result[1]
	}
	return "0"
}

func CnovaUrlToApi(url string) []string {
	cod := identifyCodProdutoCnova(url)
	return []string{
		fmt.Sprintf("http://preco.api-pontofrio.com.br/V1/Skus/PrecoVenda/?idssku=%s", cod),
		fmt.Sprintf("http://rec.pontofrio.com.br/productdetails/api/skusdetails/getbyids?ids=%s", cod),
	}

}

func lojaCnovaParaGenerico(p ProdutoCNova) ProdutoGenerico {
	produto := ProdutoGenerico{}
	produto.IDProduto = identifyCodProdutoCnova(p.Link)
	produto.Nome = p.Detalhes[0].NomeProduto
	produto.Valor = p.Valores[0].PrecoVenda.Preco
	produto.Loja = identifyNomeLoja(p.Link)
	produto.Link = p.Link
	produto.Imagens = []string{
		lojaCnovaImagemMount(p.Detalhes[0].IDImagem45, produto.IDProduto),
		lojaCnovaImagemMount(p.Detalhes[0].IDImagem90, produto.IDProduto),
		lojaCnovaImagemMount(p.Detalhes[0].IDImagem130, produto.IDProduto),
		lojaCnovaImagemMount(p.Detalhes[0].IDImagem292, produto.IDProduto),
		lojaCnovaImagemMount(p.Detalhes[0].IDImagemPadrao, produto.IDProduto),
	}
	return produto
}
