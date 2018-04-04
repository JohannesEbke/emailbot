// Package emailbot provides sidecar-based functionality to process emails from an IMAP mailbox. Each email is saved as
// a plain text file (per default in the messages/ subdirectory). Emails are only downloaded once if run repeatedly.
package emailbot

import (
	"flag"
	"fmt"
	"log"
	"os"

	imapsync "github.com/JohannesEbke/go-imap-sync"
	"github.com/howeyc/gopass"
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

// ProcessFunc specifies a function processing an email with additional sidecar data.
type ProcessFunc func(emailFile string, sidecarData SidecarData) (newRecord *Record, err error)

// Process syncs mails from the given server/mailbox and processes new Emails with newEmailFunc,
// and all synced emails with allEmailFunc.
func Process(server, username, mailbox, emailDir string, newEmailFunc, allEmailFunc ProcessFunc) error {
	password := getPassword(username, server)

	result, err := imapsync.Sync(server, username, password, mailbox, emailDir)
	if err != nil {
		return err
	}
	if newEmailFunc != nil {
		for _, email := range result.NewEmails {
			err := processMail(email, newEmailFunc)
			if err != nil {
				return err
			}
		}
	}
	if allEmailFunc != nil {
		err := processAll(emailDir, allEmailFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

// Main parses command line flags and call Process with the gathered parameters and the given ProcessFunc functions
func Main(newEmailFunc, allEmailFunc ProcessFunc) error {
	var server, username, mailbox, emailDir string
	flag.StringVar(&server, "server", "", "sync from this mail server and port (e.g. mail.example.com:993)")
	flag.StringVar(&username, "username", "", "username for logging into the mail server")
	flag.StringVar(&mailbox, "mailbox", "", "mailbox to read messages from (typically INBOX or INBOX/subfolder)")
	flag.StringVar(&emailDir, "messagesDir", "messages", "local directory to save messages in")
	flag.Parse()

	if server == "" {
		log.Println("This program copies emails from an IMAP mailbox to your computer. Usage:")
		flag.PrintDefaults()
		return fmt.Errorf("required parameters not found")
	}

	return Process(server, username, mailbox, emailDir, newEmailFunc, allEmailFunc)
}
