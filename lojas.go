package sopromocao

import (
	"fmt"
	netUrl "net/url"
	"regexp"
	"strconv"
)

const CnovaPrefixImg = "http://www.casasbahia-imagens.com.br/a/1/"

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

func IdentifyNomeLoja(url string) string {
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

func ValidUrl(url string) bool {
	_, err := netUrl.Parse(url)
	if err != nil {
		return false
	} else {
		return true
	}
}

func LojaCnovaImagemMount(id int, codProduto string) string {
	return (CnovaPrefixImg + strconv.Itoa(id) + "/" + codProduto + ".jpg")
}

// EX: http://www.casasbahia.com.br/UtilidadesDomesticas/Panelas/PanelasdePressao/Panela-de-Pressao-em-Aluminio-Polido-45L-com-Visor-6014-MTA-4645916.html?recsource=whome&rectype=w17
// EX: http://www.pontofrio.com.br/TelefoneseCelulares/AcessoriosparaCelulares/acessorioscelularesSamsung/Fone-De-Ouvido-Samsung-Earpods-Volume-Microfone-9441277.html?recsource=whome&rectype=w17
func IdentifyCodProdutoCnova(url string) string {
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
	cod := IdentifyCodProdutoCnova(url)
	return []string{
		fmt.Sprintf("http://preco.api-pontofrio.com.br/V1/Skus/PrecoVenda/?idssku=%s", cod),
		fmt.Sprintf("http://rec.pontofrio.com.br/productdetails/api/skusdetails/getbyids?ids=%s", cod),
	}

}

func LojaCnovaParaGenerico(p ProdutoCNova) ProdutoGenerico {
	produto := ProdutoGenerico{}
	produto.IDProduto = IdentifyCodProdutoCnova(p.Link)
	produto.Nome = p.Detalhes[0].NomeProduto
	produto.Valor = p.Valores[0].PrecoVenda.Preco
	produto.Loja = IdentifyNomeLoja(p.Link)
	produto.Link = p.Link
	produto.Imagens = []string{
		LojaCnovaImagemMount(p.Detalhes[0].IDImagem45, produto.IDProduto),
		LojaCnovaImagemMount(p.Detalhes[0].IDImagem90, produto.IDProduto),
		LojaCnovaImagemMount(p.Detalhes[0].IDImagem130, produto.IDProduto),
		LojaCnovaImagemMount(p.Detalhes[0].IDImagem292, produto.IDProduto),
		LojaCnovaImagemMount(p.Detalhes[0].IDImagemPadrao, produto.IDProduto),
	}
	return produto
}
