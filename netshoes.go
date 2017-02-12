package radardaoferta

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type NetshoesProduto struct {
	Name    string
	ID      string `json: "base_sku"`
	Link    string
	Message string
	Value   struct {
		Price struct {
			Atual string `json:"actual_price"`
		}
		Gallery []struct {
			Items []struct {
				Large string
			}
		}
	}
}

func NetshoesParse(url string) (ProdutoGenerico, error) {
	idUrl, err := NetshoesIdentificaIdApi(url)
	if err != nil {
		log.Println(err)
		return ProdutoGenerico{}, errors.New("nao identificou id da api netshoes")
	}
	urlApi := NetshoesMountUrl(idUrl)
	produto := NetshoesProduto{}
	Request(urlApi, &produto)
	if produto.Message != "success" {
		return ProdutoGenerico{}, errors.New("api da netshoes nao retornou dados")
	}
	produto.ID = idUrl
	produto.Link = url

	return NetshoesDeParaGenerico(produto), nil
}

func NetshoesIdentificaIdApi(url string) (string, error) {
	//http://www.netshoes.com.br/produto/chuteira-adidas-artilheira-in-futsal-D13-3054-244?&lkey=a2
	r, err := regexp.Compile("(\\w{2,5}\\-\\w{2,5}\\-\\w{2,5})(\\?|$)")
	if err != nil {
		panic(err)
	}
	idUrl := r.FindStringSubmatch(url)
	if len(idUrl) == 0 {
		return "", errors.New("nao identificou o id da url")
	}
	return idUrl[1], nil
}

func NetshoesMountUrl(idProduto string) string {
	return fmt.Sprintf("http://www.netshoes.com.br/services/get-complete-product-vo.jsp?_dyncharset=utf-8&productId=%s", idProduto)
}

func NetshoesDeParaGenerico(p NetshoesProduto) ProdutoGenerico {
	g := ProdutoGenerico{}
	g.Link = p.Link
	g.IDProduto = p.ID
	g.Nome = p.Name
	g.Imagens = []string{"http://" + p.Value.Gallery[0].Items[0].Large[2:]}
	g.Created = TimeNowIso()
	g.Loja = "netshoes"
	preco, err := strconv.ParseFloat(
		strings.Replace(p.Value.Price.Atual[3:], ",", ".", 1), 32)
	if err == nil {
		g.Valor = float32(preco)
	}
	return g
}
