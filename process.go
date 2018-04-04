package emailbot

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	flock "github.com/theckman/go-flock"
	yaml "gopkg.in/yaml.v2"
)

func processAll(path string, processFunction func(string, SidecarData) (*Record, error)) error {
	matches, err := filepath.Glob(fmt.Sprintf("%s/*.eml", path))
	if err != nil {
		return err
	}
	for _, email := range matches {
		err := processMail(email, processFunction)
		if err != nil {
			return err
		}
	}
	return nil
}

func processMail(email string, processFunction func(string, SidecarData) (*Record, error)) error {
	// First, lock the email file so that only one processMail ever runs at the same time
	fileLock := flock.NewFlock(email)
	locked, err := fileLock.TryLock()
	if err != nil {
		return fmt.Errorf("%s could not be locked: %s", email, err)
	}
	if !locked {
		return fmt.Errorf("%s is already locked - are multiple processes running?", email)
	}
	defer func() {
		if err = fileLock.Unlock(); err != nil {
			log.Printf("Error unlocking file (ignored): %s", err)
		}
	}()

	// Now, load the sidecar file if it exists
	sidecarFilename := email + ".emailbot.yaml"
	hasSidecar, err := fileExists(sidecarFilename)
	if err != nil {
		return fmt.Errorf("Error looking for %s: %s", sidecarFilename, err)
	}
	data := &SidecarData{}
	if hasSidecar {
		data, err = readSidecar(sidecarFilename)
		if err != nil {
			return fmt.Errorf("Error reading %s: %s", sidecarFilename, err)
		}
	}
	newRecord, err := processFunction(email, *data)
	if err != nil {
		return fmt.Errorf("Error processing %s: %s", email, err)
	}
	if newRecord != nil {
		data.Records = append(data.Records, *newRecord)
		err := writeSidecar(sidecarFilename, *data)
		if err != nil {
			return fmt.Errorf("Error writing %s: %s", sidecarFilename, err)
		}
	}
	return nil
}

func readSidecar(path string) (*SidecarData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	t := SidecarData{}
	err = yaml.Unmarshal(fileBytes, &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &t, nil
}

func writeSidecar(path string, data SidecarData) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0600)
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

// SidecarData represents a list of Records
type SidecarData struct {
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
