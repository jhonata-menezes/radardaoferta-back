package sopromocao

import (
	"fmt"
	netUrl "net/url"
	"regexp"
	"strconv"
)

const CnovaPrefixImg = "http://www.casasbahia-imagens.com.br/a/1/"

const GrupoCnova = "cnova"
const GrupoB2w = "b2w"

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

func IdentifyNomeLoja(url string) (string, string) {
	urlLoja, err := netUrl.Parse(url)
	if err != nil {
		panic(err)
	}
	switch urlLoja.Host {
	case "pontofrio.com.br", "extra.com.br", "casasbahia.com.br", "cdiscount.com.br", "www.pontofrio.com.br", "www.extra.com.br", "www.casasbahia.com.br", "www.cdiscount.com.br":
		return urlLoja.Host, GrupoCnova
	case "submarino.com.br", "americanas.com.br", "shoptime.com.br", "soubarato.com.br", "www.submarino.com.br", "www.americanas.com.br", "www.shoptime.com.br", "www.soubarato.com.br":
		return urlLoja.Host, GrupoB2w
	default:
		return "", ""
	}
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

func IdentifyCodProdutoB2w(url string) string {
	a, err := regexp.Compile("produto\\/([0-9]+)\\/?\\??")
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

func B2wUrlToApi(url string) string {
	cod := IdentifyCodProdutoB2w(url)
	return fmt.Sprintf("http://product-v3.soubarato.com.br/product?q=itemId:(%s)&limit=1&paymentOptionIds=CARTAO_VISA,CARTAO_SUBA_MASTERCARD,BOLETO", cod)
}

func LojaCnovaParaGenerico(p ProdutoCNova) ProdutoGenerico {
	produto := ProdutoGenerico{}
	nomeLoja, _ := IdentifyNomeLoja(p.Link)
	produto.IDProduto = IdentifyCodProdutoCnova(p.Link)
	produto.Nome = p.Detalhes[0].NomeProduto
	produto.Valor = p.Valores[0].PrecoVenda.Preco
	produto.Loja = nomeLoja
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

func LojaB2wParaGenerico(p ProdutoB2w) ProdutoGenerico {
	produto := ProdutoGenerico{}
	nomeLoja, _ := IdentifyNomeLoja(p.Link)
	produto.IDProduto = IdentifyCodProdutoB2w(p.Link)
	produto.Nome = p.Products[0].Nome
	produto.Valor = p.Products[0].Offers[0].PaymentOptions.Boleto.Price
	produto.Loja = nomeLoja
	produto.Link = p.Link
	for _, u := range p.Products[0].Imagens {
		produto.Imagens = append(produto.Imagens, u.Medium)
	}
	return produto
}
