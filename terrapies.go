package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type recipe struct {
	item        string
	ingredients []string
}

type craftingPlace struct {
	id      string
	recipes []recipe
}

func main() {
	// request and parse the front page
	resp, err := http.Get("http://terraria.gamepedia.com/Recipes")
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// Get TOC
	// Get each section and gather all the recipes one-by-one by visiting each section from TOC
	idMatcher := func(n *html.Node) bool {
		if n != nil {
			return scrape.Attr(n, "class") == "mw-headline"
		}
		return false
	}

	idMatches := scrape.FindAll(root, idMatcher)
	var ids []string
	for _, id := range idMatches {
		ids = append(ids, scrape.Attr(id, "id"))
	}
	fmt.Println("Gathered ids: ", ids)
	done := make(chan bool)
	defer close(done)

	workers := make([]<-chan string, 0)

	for _, id := range ids {
		workers = append(workers, gatherForURL("http://terraria.gamepedia.com/"+id, done))
	}

	for rec := range merge(done, workers...) {
		fmt.Println(rec)
	}
}

func merge(done <-chan bool, cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	output := func(c <-chan string) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func gatherForURL(url string, done <-chan bool) <-chan string {
	out := make(chan string, 1)
	// request and parse the front page
	go func() {
		defer close(out)
		fmt.Printf("Going to URL: %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error processing: ", url)
			return
		}
		root, err := html.Parse(resp.Body)
		if err != nil {
			fmt.Println("Error body: ", err)
			return
		}

		// Get TOC
		// Get each section and gather all the recipes one-by-one by visiting each section from TOC
		craft := func(n *html.Node) bool {
			if n != nil {
				return scrape.Attr(n, "class") == "terraria outer"
			}
			return false
		}
		craftingMatches := scrape.FindAll(root, craft)
		for _, craftList := range craftingMatches {
			innerTable := scrape.FindAllNested(craftList, scrape.ByClass("mw-redirect"))
			if len(innerTable) > 0 {
				for _, link := range innerTable {
					select {
					case out <- scrape.Attr(link, "title"):
					case <-done:
						return
					}
				}
			}
		}
	}()
	return out
}
