package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var done = false
var maxPages = 10000
var count int
var fileMu sync.Mutex
var indexMu sync.Mutex
var mu sync.Mutex
var client = &http.Client{}
var visited = make(map[string]bool)
var base = "https://en.wikipedia.org"
var filename = "links.txt"
var texts = make(map[string]map[string]bool)
var stopWords = map[string]bool{
	"the": true, "is": true, "a": true, "and": true, "of": true, "to": true,
}

type Result struct {
	URL   string
	Score float64
}

var tf = make(map[string]map[string]int)
var df = make(map[string]int)

func search(query string) {
	words := strings.Fields(strings.ToLower(query))
	scores := make(map[string]float64)
	indexMu.Lock()
	defer indexMu.Unlock()
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:\"()[]{}")
		idf := math.Log(float64(maxPages) / float64(df[word]))
		for url, tfval := range tf[word] {
			scores[url] += float64(tfval) * idf
		}
	}
	var result []Result
	for url, score := range scores {
		result = append(result, Result{URL: url, Score: score})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	for i := 0; i < min(10, len(result)); i++ {
		fmt.Println(result[i].URL, result[i].Score)
	}
}
func crawl(queue chan string, wg *sync.WaitGroup, file *os.File) {

	for link := range queue {
		/*	if done {
			wg.Done()
			fmt.Println("WG decreasing")
			return
		}*/
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
			seen := make(map[string]bool)
		tokenizerLoop:
			for {

				tokentype := tokenizer.Next()
				switch tokentype {
				case html.ErrorToken:
					break tokenizerLoop
				case html.TextToken:
					text := strings.TrimSpace(tokenizer.Token().Data)
					words := strings.Fields(text)
					indexMu.Lock()

					for _, v := range words {

						v = strings.ToLower(v)
						v = strings.Trim(v, ".,!?;:\"()[]{}")
						if stopWords[v] {
							continue
						}
						if !seen[v] {
							df[v]++
							seen[v] = true
						}
						if texts[v] == nil {
							texts[v] = make(map[string]bool)
						}

						newLink := strings.TrimRight(link, "/")
						if i := strings.Index(newLink, "#"); i != -1 {
							newLink = newLink[:i]
						}
						if tf[v] == nil {
							tf[v] = make(map[string]int)
						}
						tf[v][newLink]++
						texts[v][newLink] = true

					}
					indexMu.Unlock()
				case html.StartTagToken, html.SelfClosingTagToken:
					token := tokenizer.Token()
					if token.Data == "a" {
						for _, attr := range token.Attr {
							if attr.Key == "href" {

								newLink := attr.Val

								if len(newLink) > 0 && newLink[0] == '#' {
									continue
								}
								if len(newLink) > 0 && newLink[0] == '/' {
									newLink = base + newLink
								}
								if !strings.HasPrefix(newLink, base) {
									continue
								}
								if strings.Contains(newLink, "oldid=") ||
									strings.Contains(newLink, "printable=") ||
									strings.Contains(newLink, "action=") {
									continue
								}
								mu.Lock()
								if visited[newLink] || count >= maxPages {
									if count >= maxPages {
										done = true
									}
									mu.Unlock()

									continue

								}
								visited[newLink] = true
								count++
								mu.Unlock()
								wg.Add(1)
								select {
								case queue <- newLink:

								default:
									wg.Done()
								}

								fileMu.Lock()
								if _, err := fmt.Fprintln(file, newLink); err != nil {
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
	link := "https://en.wikipedia.org/wiki/Cricket"
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
	search("cricket")

}
