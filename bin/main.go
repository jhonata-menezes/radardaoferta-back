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
		urls <- os.Args[2]
		urls <- os.Args[3]
	}
	close(urls)
	wg.Wait()
	fmt.Println(time.Since(inicio))
}

func processador(urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urls {
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
	}

}
