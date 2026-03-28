package ledger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var ledgerPath = filepath.Join(os.Getenv("HOME"), ".config", "digby", "ledger.txt")

type Ledger struct {
	hashes map[string]bool
	file   *os.File
}

func Load() (*Ledger, error) {
	hashes := make(map[string]bool)

	err := os.MkdirAll(filepath.Dir(ledgerPath), 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	data, _ := io.ReadAll(f)

	// `\n` writes a newline in UNIX while `\r\n` writes a newline in Windows
	var lines []string
	if runtime.GOOS == "windows" {
		lines = strings.Split(string(data), "\r\n")
	} else {
		lines = strings.Split(string(data), "\n")
	}
	for i := 0; i < len(lines)-1; i++ {
		hashes[lines[i]] = true
	}

	return &Ledger{hashes: hashes, file: f}, nil
}

func (l *Ledger) Has(hash string) bool {
	return l.hashes[hash]
}

func (l *Ledger) Add(hash string) {
	l.hashes[hash] = true
	fmt.Fprintln(l.file, hash)
}

func (l *Ledger) Close() {
	l.file.Close()
}
