package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	sopromocao "bitbucket.org/jhonata-menezes/sopromocao-backend"
)

func main() {
	inicio := time.Now()
	var wg sync.WaitGroup
	urls := make(chan string, 100)
	wg.Add(1)
	go processador(urls, &wg)
	for i := 0; i < 1; i++ {
		urls <- os.Args[1]
	}
	close(urls)
	wg.Wait()
	fmt.Println(time.Since(inicio))
}

func processador(urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urls {
		nomeLoja, grupoLoja := sopromocao.IdentifyNomeLoja(url)
		fmt.Println(nomeLoja, grupoLoja)
		if grupoLoja == sopromocao.GrupoCnova {
			p := sopromocao.ProdutoCNova{}
			p.Link = url
			u := sopromocao.CnovaUrlToApi(url)
			sopromocao.Request(u[0], &p)
			sopromocao.Request(u[1], &p.Detalhes)
			if len(p.Valores) >= 1 {
				//Produto := lojaCnovaParaGenerico(p)
				//fmt.Printf("%#v", Produto)
				//key: value fmt.Printf("%#v", Produto)
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
				fmt.Printf("%#v", Produto)
				//key: value fmt.Printf("%#v", Produto)
			} else {
				log.Println("URL informada nao existe", url)
			}
		} else {
			fmt.Println("nao foi identificado o site", url)
		}
	}

}
