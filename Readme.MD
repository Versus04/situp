# situp

A lightweight web crawler written in Go that discovers and indexes links from a website.

## Overview

**situp** crawls web pages starting from a seed URL, extracts all hyperlinks, and writes discovered URLs to a file. It uses breadth-first search to explore linked pages and tracks visited links to avoid duplication.

## Features

- Crawls HTML pages and extracts links from `<a>` tags
- Handles relative URLs (paths starting with `/`)
- Tracks visited links to avoid re-crawling
- Writes discovered URLs to `links.txt`
- Browser-like User-Agent header to improve request acceptance
- Error recovery (skips failed requests and continues crawling)

## Requirements

- Go 1.18 or later
- `golang.org/x/net/html` package

## Installation

```bash
go mod tidy