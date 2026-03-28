package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"agent/ledger"
	"agent/uploader"
)

var supported = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".heic": true,
}

func isSupported(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return supported[ext]
}

func Run(folder string, queue chan string, lgr *ledger.Ledger) {
	fmt.Println("Scanning for existing photos...")
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !isSupported(path) {
			return nil
		}
		hash, err := uploader.HashFile(path)
		if err != nil {
			return nil
		}
		if !lgr.Has(hash) {
			queue <- path
		}
		return nil
	})
	fmt.Println("Scan complete. Watching for new files...")

	w, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer w.Close()

	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			w.Add(path)
		}
		return nil
	})

	var debounce = make(map[string]time.Time)
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				info, err := os.Stat(event.Name)
				if err != nil {
					continue
				}
				if info.IsDir() {
					w.Add(event.Name)
					continue
				}
				if !isSupported(event.Name) {
					continue
				}
				debounce[event.Name] = time.Now()
			}
		case <-time.After(500 * time.Millisecond):
			for path, lastSeen := range debounce {
				if time.Since(lastSeen) > 2*time.Second {
					queue <- path
					delete(debounce, path)
				}
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			fmt.Println("Watcher error:", err)
		}
	}
}