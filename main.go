package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/html"
)

var client = &http.Client{}
var filename = "links.txt"
var visited = make(map[string]bool)
var base = "https://www.cloudflare.com/learning/bots/what-is-a-web-crawler/"

func crawl(queue []string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	for len(queue) > 0 {
		link := queue[0]
		if len(link) > 0 && link[0] == '/' {
			link = base + link
		}
		queue = queue[1:]
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; situp-bot/1.0)")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		tokenizer := html.NewTokenizer(resp.Body)
	TokenizerLoop:
		for {
			tokenType := tokenizer.Next()

			switch tokenType {
			case html.ErrorToken:
				// End of the document
				break TokenizerLoop
			case html.StartTagToken, html.SelfClosingTagToken:
				token := tokenizer.Token()
				// Check if the tag is an anchor tag "a"
				if token.Data == "a" {
					for _, attr := range token.Attr {
						if attr.Key == "href" {
							link := attr.Val
							if len(link) > 0 && link[0] == '/' {
								link = base + link
							}
							if visited[link] {
								continue
							}
							queue = append(queue, link)
							visited[link] = true
							if _, err := fmt.Fprintln(file, link); err != nil {
								continue
							}

						}
					}
				}
			}
		}
		resp.Body.Close()
	}

}
func main() {
	//c := make(chan string)
	queue := []string{"https://en.wikiversity.org/wiki/Wikiversity:Main_Page"}
	crawl(queue)
}
