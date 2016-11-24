package sopromocao

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
)

type NetshoesProduto struct {
	ID      string
	Link    string
	Message string
	Value   struct {
		Ecommerce struct {
			Detail struct {
				Products []struct {
					Id          string
					price       string
					Dimension30 string
				}
			}
		}
		Chaordic struct {
			Product struct {
				Name  string
				Price float32
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
	urlApi := NetshoesIdentificaUrlApi(idUrl)

	produto := NetshoesProduto{}
	Request(urlApi, &produto)
	produto.ID = idUrl
	produto.Link = url

	return NetshoesDeParaGenerico(produto), nil
}

func NetshoesIdentificaIdApi(url string) (string, error) {
	//http://www.netshoes.com.br/produto/chuteira-adidas-artilheira-in-futsal-D13-3054-244?&lkey=a2
	r, err := regexp.Compile("(\\w{2,5}\\-\\w{2,5}\\-\\w{2,5})\\?")
	if err != nil {
		panic(err)
	}
	idUrl := r.FindStringSubmatch(url)
	if len(idUrl) == 0 {
		return "", errors.New("nao identificou o id da url")
	}
	return idUrl[1], nil
}

func NetshoesIdentificaUrlApi(id string) string {
	for i := 50; i >= 0; i-- {
		produto := NetshoesProduto{}
		url := NetshoesMountUrl(id, strconv.Itoa(i))
		Request(url, &produto)
		if produto.Message == "success" {
			return url
		}
	}
	return ""
	//http://www.netshoes.com.br/services/analytics-product-data.jsp?productId=D13-3054&skuId=D13-3054-244-44

}

func NetshoesMountUrl(idProduto string, idTentativa string) string {
	r, err := regexp.Compile("(\\w{2,5}\\-\\w{2,5})\\-")
	if err != nil {
		panic(err)
	}
	subId := r.FindStringSubmatch(idProduto)
	return fmt.Sprintf("http://www.netshoes.com.br/services/analytics-product-data.jsp?productId=%s&skuId=%s-%s",
		subId[1],
		idProduto,
		idTentativa)
}

func NetshoesImagem(idUrl string) string {
	//http://static1.netshoes.net/Produtos/chuteira-adidas-artilheira-in-futsal/44/D13-3054-244/D13-3054-244_zoom1.jpg?resize=544:*
	return fmt.Sprintf("http://static1.netshoes.net/Produtos/chuteira-adidas-artilheira-in-futsal/%s/%s/%s_zoom1.jpg?resize=400:*", idUrl[len(idUrl)-2:], idUrl, idUrl)
}

func NetshoesDeParaGenerico(p NetshoesProduto) ProdutoGenerico {
	g := ProdutoGenerico{}
	g.Link = p.Link
	g.IDProduto = p.ID
	g.Nome = p.Value.Chaordic.Product.Name
	g.Valor = p.Value.Chaordic.Product.Price
	g.Imagens = []string{NetshoesImagem(p.ID)}
	g.Created = TimeNowIso()
	g.Loja = "netshoes"
	return g
}
