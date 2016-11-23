package main

//key: value fmt.Printf("%#v", Produto)

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"flag"

	sopromocao "bitbucket.org/jhonata-menezes/sopromocao-backend"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var produtosJson []byte
var produtosCollection []sopromocao.ProdutoGenerico
var chanUrls chan string
var connMongo *mgo.Session
var limitProdutos = 50

func main() {
	var host string
	var port string
	flag.StringVar(&host, "host", "127.0.0.1", "interface")
	flag.StringVar(&port, "port", "5001", "porta")
	flag.Parse()
	var wg sync.WaitGroup
	chanUrls = make(chan string, 400)
	wg.Add(1)
	go processador(chanUrls, &wg)
	go sopromocao.ShowTelegram(chanUrls)
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
	err = coll.Find(bson.M{}).Limit(limitProdutos).Sort("-_id").All(&produtosCollection)
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
	router.HandleFunc("/api/produtos/redirecionar/{id}", getRedirecionar).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(http404)

	fmt.Println("GO! http://" + host + ":" + port)
	http.ListenAndServe(host+":"+port, router)
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
				Produto.Created = time.Now().Format("2006-01-02 15:04:05")
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
				Produto.Created = time.Now().Format("2006-01-02 15:04:05")
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
		w.Write([]byte("{\"status\": \"error\", \"msg\":\"corpo da solicitacao vazio\"}"))
		http.Error(w, "", 400)
		return
	}
	var bodyUrl struct {
		Url string
	}
	err := json.NewDecoder(r.Body).Decode(&bodyUrl)
	if err != nil {
		w.Write([]byte("{\"status\": \"error\", \"msg\":\"json invalido\"}"))
		http.Error(w, "", 400)
		return
	}
	bodyUrl.Url = sopromocao.CleanUrl(bodyUrl.Url)
	urlValidator, _ := sopromocao.IdentifyNomeLoja(bodyUrl.Url)
	if urlValidator == "" {
		w.Write([]byte("{\"status\":\"error\", \"msg\": \"URL Invalida, por favor informe apenas urls de ecommerce da lista\"}"))
		return
	}

	chanUrls <- bodyUrl.Url

	w.Write([]byte("{\"status\":\"ok\"}"))
}

func getRedirecionar(w http.ResponseWriter, r *http.Request) {
	idProdutoMongo := mux.Vars(r)["id"]
	if len(idProdutoMongo) != 12 {
		http.Error(w, "", 404)
		return
	}
	produtosStruct := sopromocao.ProdutoGenerico{}

	collection := produtosColl()
	query := collection.FindId(bson.ObjectIdHex(idProdutoMongo))
	count, err := query.Count()
	if count == 0 {
		http.Error(w, "", 404)
		return
	}
	query.One(&produtosStruct)
	w.Header().Set("Location", produtosStruct.Link)
	w.Write([]byte(""))

	err = collection.UpdateId(bson.ObjectIdHex(idProdutoMongo), bson.M{"$set": bson.M{"cliques": (produtosStruct.Cliques + 1)}})
	if err != nil {
		log.Println(err)
	}
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
		err = coll.Find(bson.M{}).Limit(limitProdutos).Sort("-_id").All(&produtosCollection)
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}
