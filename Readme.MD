# situp

A concurrent Go web crawler that starts from a seed URL, discovers links from HTML anchor tags, and saves discovered links to a text file.

## Overview

This project uses a buffered channel as a work queue and multiple crawler goroutines to fetch pages in parallel. Each page is parsed with the HTML tokenizer from golang.org/x/net/html, and discovered links are added to a shared visited map to reduce duplicate processing.

## Current Behavior

- Starts from this seed URL:
	- https://en.wikibooks.org/wiki/Department:Engineering
- Uses 10 crawler workers
- Uses a queue channel with capacity 1000
- Writes discovered links to links.txt
- Adds a browser-like User-Agent header for requests
- Skips href values that start with:
	- javascript:
	- mailto:
	- tel:
	- # (fragment links)
- Converts links that start with / using this base:
	- https://en.wikibooks.org/

## Requirements

- Go 1.18 or newer

## Setup

```bash
go mod tidy
```

## Run

```bash
go run .
```

When the run completes, the crawler prints Done and discovered links are appended to links.txt.

## Notes

- The crawler appends output to links.txt on every run.
- Relative URL handling currently only converts links beginning with /.
- Links are enqueued with a non-blocking send; when the queue is full, new discovered links are dropped.

## Project Files

- main.go: crawler implementation
- go.mod, go.sum: module and dependency metadata
- links.txt: discovered links output