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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var produtosJson []byte
var produtosCollection []sopromocao.ProdutoGenerico
var chanUrls chan string
var connMongo *mgo.Session

func main() {
	var wg sync.WaitGroup
	chanUrls = make(chan string, 400)
	wg.Add(1)
	go processador(chanUrls, &wg)
	//close(urls)
	//wg.Wait()

	c, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	defer c.Close()
	c.SetMode(mgo.Monotonic, true)

	connMongo = c

	coll := produtosColl()
	err = coll.Find(bson.M{}).Limit(30).All(&produtosCollection)
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(produtosCollection)
	if err != nil {
		panic(err)
	}
	produtosJson = json

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("Content-Type")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, content-type, x-correlation-id, Origin, Host, User-Agent, Access-Control-Request-Headers, Referer, Connection, Accept, Accept-Language, Access-Control-Request-Method, Accept-Encoding")
		//responseDefault(w)
	})
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
	responseDefault(w)
	w.Write(produtosJson)
}

func postNovoProduto(w http.ResponseWriter, r *http.Request) {
	responseDefault(w)
	if r.Body == nil {
		http.Error(w, "corpo da solicitacao vazio", 400)
		return
	}
	var bodyUrl struct {
		Url string
	}
	err := json.NewDecoder(r.Body).Decode(&bodyUrl)
	if err != nil {
		http.Error(w, "json invalido", 400)
		return
	}

	chanUrls <- bodyUrl.Url

	w.Write([]byte("{\"status\":\"ok\"}"))
}

func http404(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", 404)
}

func mesclaGenericoParaJSON(p sopromocao.ProdutoGenerico) {
	coll := produtosColl()
	cnt, err := coll.Find(bson.M{"loja": p.Loja, "idProduto": p.IDProduto}).Count()
	if err != nil {
		panic(err)
	}
	if cnt == 0 {
		err := coll.Insert(p)
		if err != nil {
			panic(err)
		}
		err = coll.Find(bson.M{}).Limit(30).All(&produtosCollection)
		if err != nil {
			panic(err)
		}
		json, err := json.Marshal(produtosCollection)
		if err != nil {
			panic(err)
		}
		produtosJson = json
	}
}

func produtosColl() *mgo.Collection {
	conn := connMongo.Copy()
	return conn.DB("sopromocao").C("produtos")
}

func responseDefault(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}
