package main

//key: value fmt.Printf("%#v", Produto)

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	sopromocao "bitbucket.org/jhonata-menezes/sopromocao-backend"
	"github.com/gorilla/mux"
)

var produtosJson []byte
var produtosCollection []sopromocao.ProdutoGenerico
var chanUrls chan string

func main() {
	var wg sync.WaitGroup
	chanUrls = make(chan string, 400)
	wg.Add(1)
	go processador(chanUrls, &wg)
	//close(urls)
	//wg.Wait()

	json, _ := json.Marshal(sopromocao.ProdutoGenerico{})
	produtosJson = json

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/api/produtos", getProduto).Methods("GET")
	router.HandleFunc("/api/produtos/novo", postNovoProduto).Methods("POST")
	router.NotFoundHandler = http.HandlerFunc(http404)

	fmt.Println("GO!")
	http.ListenAndServe("0.0.0.0:5014", router)
}

func processador(urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var nomeLoja, grupoLoja string

	for url := range urls {
		nomeLoja, grupoLoja = sopromocao.IdentifyNomeLoja(url)
		fmt.Println(nomeLoja, grupoLoja)
		if grupoLoja == sopromocao.GrupoCnova {
			p := sopromocao.ProdutoCNova{}
			p.Link = url
			u := sopromocao.CnovaUrlToApi(url)
			sopromocao.Request(u[0], &p)
			sopromocao.Request(u[1], &p.Detalhes)
			if len(p.Valores) >= 1 {
				Produto := sopromocao.LojaCnovaParaGenerico(p)
				mesclaGenericoParaJSON(Produto)
			} else {
				log.Println("URL informada nao existe", url)
			}
		} else if grupoLoja == sopromocao.GrupoB2w {
			p := sopromocao.ProdutoB2w{}
			p.Link = url
			u := sopromocao.B2wUrlToApi(url)
			sopromocao.Request(u, &p)
			if len(p.Products) >= 1 {
				Produto := sopromocao.LojaB2wParaGenerico(p)
				mesclaGenericoParaJSON(Produto)
			} else {
				log.Println("URL informada nao existe", url)
			}
		} else {
			fmt.Println("nao foi identificado o site", url)
		}
	}
}

func getProduto(w http.ResponseWriter, r *http.Request) {
	w.Write(produtosJson)
}

func postNovoProduto(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	chanUrls <- url

	w.Write([]byte("{\"status\":\"ok\"}"))
}

func http404(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("404"))
}

func mesclaGenericoParaJSON(p sopromocao.ProdutoGenerico) {
	//fmt.Printf("%#v", p)
	produtosCollection = append(produtosCollection, p)
	//fmt.Printf("%#v", produtosCollection)
	json, err := json.Marshal(produtosCollection)
	if err != nil {
		panic(err)
	}
	produtosJson = json
}
