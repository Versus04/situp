# situp

situp is a concurrent Go crawler with a simple inverted index for keyword search.

## What It Does

- Crawls pages concurrently using worker goroutines and a buffered queue.
- Extracts links from anchor tags (`<a href="...">`) and crawls discovered pages.
- Extracts text tokens from pages and builds an in-memory word-to-URL index.
- Supports a basic `search(query)` over crawled pages.
- Appends discovered links to `links.txt`.

## Current Runtime Configuration

Based on the current `main.go`:

- Seed URL: `https://en.wikipedia.org/wiki/Cricket`
- Worker count: `10`
- Queue buffer size: `1000`
- Crawl page limit (`maxPages`): `10000`
- Output file: `links.txt`

## Crawl and Index Logic

- Requests use a browser-like User-Agent.
- Non-200 responses are skipped.
- `href` values are filtered:
  - Skips `javascript:`, `mailto:`, `tel:`, and fragment-only links (`#...`).
  - Converts links starting with `/` using `base`.
  - Keeps only links that start with `base`.
- Duplicate links are prevented using `visited`.
- Text content is tokenized, normalized to lowercase, punctuation-trimmed, and indexed.
- Stop words are ignored during indexing.

## Search

After crawling, `main()` runs:

```go
result := search("cricket")
```

`search(query)`:

- Splits query into words.
- Looks up matching URLs from the in-memory index (`texts`).
- Returns a list of URLs that contain one or more query words.

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
  - Search result URLs
  - `Done` when execution completes
- File:
  - Discovered links appended to `links.txt`

## Important Notes

- The crawler uses shared global state (`visited`, `texts`, `count`, `done`) protected by mutexes.
- The domain filter uses `base`. In current code, `base` is set to `https://en.wikibooks.org/` while the seed URL is on `wikipedia.org`, so discovered links outside `base` are intentionally ignored.
- `links.txt` is append-only; delete it manually if you want a fresh run.

## Project Files

- `main.go`: crawler + index + search flow
- `go.mod`, `go.sum`: module metadata
- `links.txt`: discovered URL output