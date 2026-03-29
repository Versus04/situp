package crawler

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Versus04/situp/index"
	"golang.org/x/net/html"
)

type Config struct {
	BaseURL     string
	MaxPages    int
	WorkerCount int
	QueueSize   int
	UserAgent   string
	OutputFile  string
}

type Crawler struct {
	cfg     Config
	client  *http.Client
	visited map[string]bool
	count   int

	mu     sync.Mutex
	fileMu sync.Mutex
}

func New(cfg Config, client *http.Client) *Crawler {
	if client == nil {
		client = &http.Client{}
	}
	return &Crawler{
		cfg:     cfg,
		client:  client,
		visited: make(map[string]bool),
	}
}

func canonicalizeForIndexing(u string) string {
	u = strings.TrimRight(u, "/")
	if i := strings.Index(u, "#"); i != -1 {
		u = u[:i]
	}
	return u
}

func (c *Crawler) Run(seed string, idx *index.Index) error {
	queue := make(chan string, c.cfg.QueueSize)

	file, err := os.OpenFile(c.cfg.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	queue <- seed

	for i := 0; i < c.cfg.WorkerCount; i++ {
		go c.worker(queue, &wg, file, idx)
	}

	go func() {
		wg.Wait()
		close(queue)
	}()

	wg.Wait()
	return nil
}

func (c *Crawler) worker(queue chan string, wg *sync.WaitGroup, file *os.File, idx *index.Index) {
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
			req.Header.Set("User-Agent", c.cfg.UserAgent)

			resp, err := c.client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			seen := make(map[string]bool)
			tokenizer := html.NewTokenizer(resp.Body)

		tokenizerLoop:
			for {
				switch tokenizer.Next() {
				case html.ErrorToken:
					break tokenizerLoop

				case html.TextToken:
					text := strings.TrimSpace(tokenizer.Token().Data)
					if text == "" {
						continue
					}
					words := strings.Fields(text)
					docURL := canonicalizeForIndexing(link)
					for _, w := range words {
						idx.AddToken(docURL, w, seen)
					}

				case html.StartTagToken, html.SelfClosingTagToken:
					token := tokenizer.Token()
					if token.Data != "a" {
						continue
					}
					for _, attr := range token.Attr {
						if attr.Key != "href" {
							continue
						}

						newLink := attr.Val
						if len(newLink) > 0 && newLink[0] == '#' {
							continue
						}
						if len(newLink) > 0 && newLink[0] == '/' {
							newLink = c.cfg.BaseURL + newLink
						}
						if !strings.HasPrefix(newLink, c.cfg.BaseURL) {
							continue
						}
						if strings.Contains(newLink, "oldid=") ||
							strings.Contains(newLink, "printable=") ||
							strings.Contains(newLink, "action=") {
							continue
						}

						c.mu.Lock()
						if c.visited[newLink] || c.count >= c.cfg.MaxPages {
							c.mu.Unlock()
							continue
						}
						c.visited[newLink] = true
						c.count++
						c.mu.Unlock()

						wg.Add(1)
						select {
						case queue <- newLink:
						default:
							wg.Done()
						}

						c.fileMu.Lock()
						_, _ = fmt.Fprintln(file, newLink)
						c.fileMu.Unlock()
					}
				}
			}
		}()
	}
}
