package sopromocao

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type WalmartProduto struct {
	Product []struct {
		ID      int    `json:"productId"`
		Nome    string `json:"productName"`
		Valor   string `json:"productPrice"`
		Imagens string `json:"productImage"`
	}
	Link string
}

func WalmartParse(url string) (ProdutoGenerico, error) {
	produto, err := WalmartGetInfo(url)
	if err != nil {
		return ProdutoGenerico{}, err
	}
	produto.Link = url
	return WalmartDePara(produto), nil
}

func WalmartGetInfo(url string) (WalmartProduto, error) {
	url = strings.Replace(url, "://", "://www.", 1)
	body := string(RequestBody(url))
	r, _ := regexp.Compile("\\<script\\>var dataLayer \\= \\[(.*?)\\]\\;dataLayer\\.push")
	match := r.FindStringSubmatch(string(body))
	if len(match) >= 2 {
		produto := WalmartProduto{}
		err := json.Unmarshal([]byte(match[1]), &produto)
		if err != nil {
			return WalmartProduto{}, err
		}
		return produto, nil
	}
	return WalmartProduto{}, errors.New("walmart nao identificou o javascript")
}

func WalmartDePara(w WalmartProduto) ProdutoGenerico {
	p := ProdutoGenerico{}
	p.IDProduto = strconv.Itoa(w.Product[0].ID)
	p.Nome = w.Product[0].Nome
	p.Created = TimeNowIso()
	p.Loja = "walmart"

	valor64, err := strconv.ParseFloat(w.Product[0].Valor, 32)
	if err == nil {
		p.Valor = float32(valor64)
	}
	imagem := "http://" + strings.Replace(w.Product[0].Imagens[2:], ".jpg", "/a.jpg", 1)
	p.Imagens = []string{imagem}
	p.Link = strings.Replace(w.Link, "://", "://www.", 1)
	return p
}
