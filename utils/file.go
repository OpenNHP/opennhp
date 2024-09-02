package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/OpenNHP/opennhp/log"
	"github.com/fsnotify/fsnotify"
)

func ReadWholeFile(fileName string) (string, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", err
	}

	buf, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func HashFile(method string, fileName string) (string, error) {
	file, err := os.Open(fileName) // Open the file for reading
	if err != nil {
		return "", err
	}
	defer file.Close() // Be sure to close your file

	var hash hash.Hash

	switch {
	case strings.EqualFold(method, "md5"):
		hash = md5.New()

	case strings.EqualFold(method, "sha1"):
		hash = sha1.New()

	case strings.EqualFold(method, "sha256"):
		hash = sha256.New()
	}

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil // Get hex encoded hash sum
}

type fileWatcher struct {
	filename string
	watcher  *fsnotify.Watcher
	wait     *sync.WaitGroup
}

func (w *fileWatcher) Close() error {
	w.watcher.Close()
	w.wait.Wait()
	log.Info("file watcher for %s closed", w.filename)
	return nil
}

func WatchFile(file string, callback func()) io.Closer {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("failed to create file watcher for %s: %v", file, err)
		return nil
	}

	// we have to watch the entire directory to pick up renames/atomic saves in a cross-platform way
	filename := filepath.Clean(file)
	dirPath, _ := filepath.Split(filename)
	watcher.Add(dirPath)

	var eventsWG sync.WaitGroup
	var debounceTimer *time.Timer
	debounceTime := 100 * time.Millisecond
	eventsWG.Add(1)

	go func() {
		defer CatchPanic()
		defer watcher.Close()
		defer eventsWG.Done()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok { // 'Events' channel is closed
					return
				}

				// callback fires when file is modified or created
				if filepath.Clean(event.Name) == filename {
					if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
						if callback != nil {
							if debounceTimer != nil {
								debounceTimer.Stop()
							}
							debounceTimer = time.AfterFunc(debounceTime, func() { callback() })
						}
					} else if event.Has(fsnotify.Remove) {
						return
					}
				}

			case err, ok := <-watcher.Errors:
				if ok { // 'Errors' channel is not closed
					log.Error(fmt.Sprintf("file watcher error: %v", err))
				}
				return
			}
		}
	}()

	log.Info("start watching file %s", filename)
	return &fileWatcher{filename, watcher, &eventsWG}
}
