package sopromocao

import (
	"errors"
	"fmt"
	"regexp"
)

type MagazineLuizaProduto struct {
	Product struct {
		ID        string   `json:"id"`
		Descricao string   `json:"complete_description"`
		Preco     float32  `json:"discount_price"`
		Imagens   []string `json:"images"`
		Nome      string   `json:"title"`
	}
	Link string
}

func MagazineLuizaParse(url string) (ProdutoGenerico, error) {
	id, err := MagazineLuizaGetID(url)
	if err != nil {
		return ProdutoGenerico{}, err
	}
	produtoMagazine := MagazineLuizaGetProduto(id)
	produtoMagazine.Link = url
	return MagazineLuizaDePara(produtoMagazine), nil
}

func MagazineLuizaGetID(url string) (string, error) {
	r, _ := regexp.Compile("p\\/(\\d+)\\/")
	idProduto := r.FindStringSubmatch(url)
	if len(idProduto) == 0 {
		return "", errors.New("nao identificou id da magazine luiza")
	}
	return idProduto[1], nil
}

func MagazineLuizaGetProduto(id string) MagazineLuizaProduto {
	//https://m.magazineluiza.com.br/catalog/products/2160329.json
	produtoMagazine := MagazineLuizaProduto{}
	Request(fmt.Sprintf("https://m.magazineluiza.com.br/catalog/products/%s.json", id), &produtoMagazine)
	return produtoMagazine
}

func MagazineLuizaDePara(m MagazineLuizaProduto) ProdutoGenerico {
	p := ProdutoGenerico{}
	p.IDProduto = m.Product.ID
	p.Created = TimeNowIso()
	p.Valor = m.Product.Preco
	p.Link = m.Link
	p.Loja = "Magazine Luiza"
	p.Nome = m.Product.Nome
	if len(m.Product.Imagens) >= 1 {
		p.Imagens = []string{mountImagem(m.Product.Imagens[0])}
	}
	return p
}

func mountImagem(idImagem string) string {
	return fmt.Sprintf("http://c.mlcdn.com.br/470x352/%s", idImagem)
}
