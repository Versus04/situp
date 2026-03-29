package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Versus04/situp/crawler"
	"github.com/Versus04/situp/index"
	"github.com/Versus04/situp/search"
)

func main() {
	cfg := crawler.Config{
		BaseURL:     "https://en.wikipedia.org",
		MaxPages:    10000,
		WorkerCount: 10,
		QueueSize:   1000,
		UserAgent:   "Mozilla/5.0 (compatible; situp-bot/1.0)",
		OutputFile:  "links.txt",
	}

	seed := "https://en.wikipedia.org/wiki/Cricket"

	idx := index.NewDefault()
	c := crawler.New(cfg, &http.Client{})

	if err := c.Run(seed, idx); err != nil {
		log.Fatal(err)
	}

	for _, r := range search.Query(idx, "cricket", cfg.MaxPages, 10) {
		fmt.Println(r.URL, r.Score)
	}
}
