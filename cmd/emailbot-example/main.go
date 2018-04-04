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
