[![GoDoc](https://godoc.org/github.com/JohannesEbke/emailbot?status.svg)](https://godoc.org/github.com/JohannesEbke/emailbot) [![Go Report Card](https://goreportcard.com/badge/github.com/JohannesEbke/emailbot)](https://goreportcard.com/report/github.com/JohannesEbke/emailbot)

# emailbot
Library to develop email-based bots.

## What does it do?
This library syncs an IMAP folder to disk, and then calls processing functions separately on new and all emails.
It can record its steps in YAML sidecar files next to each email. It uses `flock` to lock processing steps.

## Based on
[github.com/JohannesEbke/go-imap-sync](https://github.com/JohannesEbke/go-imap-sync)

## Example (complete)
```
package main

import (
	"log"
	"time"
	"github.com/JohannesEbke/emailbot"
)

func main() {
	emailbot.Main(addDownloadedRecord, printDetails)
}

func addDownloadedRecord(_ string, data emailbot.SidecarData) (*emailbot.Record, error) {
	return &emailbot.Record{Time: time.Now(), Key: "synced"}, nil
}

func printDetails(emailFile string, data emailbot.SidecarData) (*emailbot.Record, error) {
	log.Printf("%s: %v", emailFile, data)
	return nil, nil
}
```
