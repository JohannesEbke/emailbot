// go-imap-sync provides a simple command line tool to download emails from an IMAP mailbox. Each email is saved as a
// plain text file (per default in the messages/ subdirectory). Emails are only downloaded once if run repeatedly.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/JohannesEbke/go-imap-sync"
	"github.com/howeyc/gopass"
	flock "github.com/theckman/go-flock"
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
			fileLock := flock.NewFlock(email)
			locked, err := fileLock.TryLock()
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if err := fileLock.Unlock(); err != nil {
					log.Fatal(err)
				}
			}()
			if locked {
				data, err := readSidecar(sidecarFilename)
				time.Sleep(10000000000)
				log.Println("sleep")
				if err != nil {
					log.Fatal(err)
				}
				log.Println(data)
				writeSidecar(sidecarFilename, *data)
			} else {
				log.Fatal("File could not be Locked!")
			}
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

func writeSidecar(path string, data SidecarFile) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

// Record of an automatic action attempt or result
type Record struct {
	// Time the Record was inserted
	Time time.Time `yaml:"time"`
	// Key of the Record
	Key string `yaml:"key"`
	// Output data of the Record
	Data string `yaml:"data,omitempty"`
}

// SidecarFile represents a list of Records
type SidecarFile struct {
	Records []Record `yaml:"process_records"`
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
