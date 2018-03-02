// go-imap-sync provides a simple command line tool to download emails from an IMAP mailbox. Each email is saved as a
// plain text file (per default in the messages/ subdirectory). Emails are only downloaded once if run repeatedly.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/JohannesEbke/go-imap-sync"
	"github.com/howeyc/gopass"
	yaml "gopkg.in/yaml.v2"
)

func getPassword(username, server string) (password string) {
	password = os.Getenv("IMAP_PASSWORD")

	if password == "" {
		log.Printf("Enter IMAP Password for %v on %v: ", username, server)
		passwordBytes, err := gopass.GetPasswd()
		if err != nil {
			panic(err)
		}
		password = string(passwordBytes)
	}
	return
}

func main() {
	var server, username, mailbox, emailDir string
	flag.StringVar(&server, "server", "", "sync from this mail server and port (e.g. mail.example.com:993)")
	flag.StringVar(&username, "username", "", "username for logging into the mail server")
	flag.StringVar(&mailbox, "mailbox", "", "mailbox to read messages from (typically INBOX or INBOX/subfolder)")
	flag.StringVar(&emailDir, "messagesDir", "messages", "local directory to save messages in")
	flag.Parse()

	if server == "" {
		log.Println("emailbot copies emails from an IMAP mailbox to your computer. Usage:")
		flag.PrintDefaults()
		log.Fatal("Required parameters not found.")
	}

	password := getPassword(username, server)

	result, err := imapsync.Sync(server, username, password, mailbox, emailDir)
	if err != nil {
		log.Fatal(err)
	}
	allEmails := append(result.ExistingEmails, result.NewEmails...)
	for _, email := range allEmails {
		log.Println(email)
		sidecarFilename := email + ".emailbot.yaml"
		hasSidecar, err := fileExists(sidecarFilename)
		if err != nil {
			log.Fatal(err)
		}
		if hasSidecar {
			log.Println(readSidecar(sidecarFilename))
		}
	}

}

func readSidecar(path string) (*SidecarFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	t := SidecarFile{}
	err = yaml.Unmarshal(fileBytes, &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &t, nil
}

type Step struct {
	Name string `yaml:"name"`
}

type SidecarFile struct {
	Steps []Step `yaml:"process_steps,flow"`
}

// fileExists checks if the given path exists and can be Stat'd.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
