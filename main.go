package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var fileMu sync.Mutex

var mu sync.Mutex
var client = &http.Client{}
var visited = make(map[string]bool)
var base = "https://en.wikibooks.org/"
var filename = "links.txt"

func crawl(queue chan string, wg *sync.WaitGroup, file *os.File) {

	for link := range queue {
		func() {
			defer wg.Done()
			if strings.HasPrefix(link, "javascript:") ||
				strings.HasPrefix(link, "mailto:") ||
				strings.HasPrefix(link, "tel:") {
				return
			}
			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; situp-bot/1.0)")
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return
			}
			tokenizer := html.NewTokenizer(resp.Body)

		tokenizerLoop:
			for {
				tokentype := tokenizer.Next()
				switch tokentype {
				case html.ErrorToken:
					break tokenizerLoop
				case html.StartTagToken, html.SelfClosingTagToken:
					token := tokenizer.Token()
					if token.Data == "a" {
						for _, attr := range token.Attr {
							if attr.Key == "href" {
								link := attr.Val
								if len(link) > 0 && link[0] == '#' {
									continue
								}
								if len(link) > 0 && link[0] == '/' {
									link = base + link
								}
								mu.Lock()
								if visited[link] {
									mu.Unlock()
									continue

								}
								visited[link] = true
								mu.Unlock()

								select {
								case queue <- link:
									// success
									wg.Add(1)
								default:
									// queue full → drop or skip
								}

								fileMu.Lock()
								if _, err := fmt.Fprintln(file, link); err != nil {
									fileMu.Unlock()
									continue
								}
								fileMu.Unlock()
							}
						}
					}
				}
			}

			resp.Body.Close()
		}()
	}

}
func main() {
	c := make(chan string, 1000)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	link := "https://en.wikibooks.org/wiki/Department:Engineering"
	var wg sync.WaitGroup
	wg.Add(1)
	c <- link
	for i := 0; i < 10; i++ {
		go crawl(c, &wg, file)
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	wg.Wait()
	fmt.Println("Done")

}
