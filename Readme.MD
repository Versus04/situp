# situp (Search It Up)

situp ("Search It Up") is a concurrent Go crawler with TF-IDF based keyword search.

## What It Does

- Crawls pages concurrently using worker goroutines and a buffered queue.
- Extracts links from anchor tags (`<a href="...">`) and crawls discovered pages.
- Extracts text tokens from pages and builds in-memory term statistics.
- Supports `search(query)` ranked by TF-IDF score.
- Appends discovered links to `links.txt`.

## Current Runtime Configuration

Based on the current `main.go`:

- Seed URL: `https://en.wikipedia.org/wiki/Cricket`
- Worker count: `10`
- Queue buffer size: `1000`
- Crawl page limit (`maxPages`): `10000`
- Output file: `links.txt`
- Domain base filter: `https://en.wikipedia.org`

## Crawl and Index Logic

- Requests use a browser-like User-Agent.
- Non-200 responses are skipped.
- `href` values are filtered:
  - Skips `javascript:`, `mailto:`, `tel:`, and fragment-only links (`#...`).
  - Converts links starting with `/` using `base`.
  - Keeps only links that start with `base`.
  - Skips links containing `oldid=`, `printable=`, or `action=`.
- Duplicate links are prevented using `visited`.
- Text content is tokenized, normalized to lowercase, punctuation-trimmed, and indexed.
- Stop words are ignored.
- Index data maintained in memory:
  - `TF[term][url]` term frequency per URL
  - `DF[term]` document frequency per term

## Search

After crawling, `main()` runs:

```go
search("cricket")
```

`search.Query(...)`:

- Splits query into words.
- Computes score per URL as `tf * idf` (summed across query terms).
- Uses `idf = log(maxPages / df[word])`.
- Sorts by score descending and prints up to top 10 results.

## Requirements

- Go 1.18+
- Module dependency: `golang.org/x/net/html`

## Setup

```bash
go mod tidy
```

## Run

```bash
go run .
```

## Output

- Console:
  - Top ranked URLs with score from `search("cricket")`
- File:
  - Discovered links appended to `links.txt`

## Important Notes

- The crawler and index are in-memory and concurrency-safe (mutex-protected).
- The domain filter is currently aligned with the seed domain (`wikipedia.org`).
- `links.txt` is append-only; delete it manually if you want a fresh run.

## Project Files

- `main.go`: wires crawler + index + search together
- `crawler/`: crawling, link extraction, and writing discovered links
- `index/`: token normalization, stop-words, and TF/DF storage
- `search/`: TF-IDF scoring over the index
- `go.mod`, `go.sum`: module metadata
- `links.txt`: discovered URL output