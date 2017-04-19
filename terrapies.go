package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

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
	// grab all articles and print them
	idMatches := scrape.FindAll(root, idMatcher)
	var ids []string
	for _, id := range idMatches {
		ids = append(ids, scrape.Attr(id, "id"))
	}
	fmt.Println("Gathered ids: ", ids)
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go gatherForURL("http://terraria.gamepedia.com/"+id, &wg)
	}
	wg.Wait()
}

func gatherForURL(url string, wg *sync.WaitGroup) {
	// request and parse the front page
	fmt.Printf("Going to URL: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error processing: ", url)
		wg.Done()
		return
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error body: ", err)
		wg.Done()
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
		fmt.Println("Craftlist for URL: ", url)
		fmt.Println(craftList)
	}
	wg.Done()
}
