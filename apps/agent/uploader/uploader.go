package uploader

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"

	"agent/ledger"
)

type metadata struct {
	imageID       string
	agentID       string
	fileName      string
	fileType      string
	fileTimestamp time.Time
	tags          []string
}

func Run(agentID string, watchFolder string, queue chan string, l *ledger.Ledger) {
	for path := range queue {
		err := process(path, agentID, watchFolder, l)
		if err != nil {
			fmt.Printf("Failed to process %s: %v\n", filepath.Base(path), err)
		}
	}
}

func process(path string, agentID string, watchFolder string, l *ledger.Ledger) error {
	hash, err := HashFile(path)
	if err != nil {
		return fmt.Errorf("Hashing failed: %w", err)
	}

	if l.Has(hash) {
		fmt.Printf("Skipping already uploaded: %s\n", filepath.Base(path))
		return nil
	}

	meta, err := buildMetadata(path, hash, agentID, watchFolder)
	if err != nil {
		return fmt.Errorf("Could not build metadata: %w", err)
	}

	var uploadErr error
	for attempt := 1; attempt <= 3; attempt++ {
		uploadErr = upload(path, meta)
		if uploadErr == nil {
			break
		}
		fmt.Printf("Attempt %d failed for %s: %v\n", attempt, filepath.Base(path), uploadErr)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}
	}

	if uploadErr != nil {
		return fmt.Errorf("All attempts failed: %w", uploadErr)
	}

	l.Add(hash)
	fmt.Printf("Uploaded: %s\n", filepath.Base(path))
	return nil
}

func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func buildMetadata(path, hash, agentID, watchFolder string) (metadata, error) {
	relPath, err := filepath.Rel(watchFolder, path)
	if err != nil {
		return metadata{}, fmt.Errorf("Could not get relative path: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var fileType string
	switch ext {
	case ".jpg", ".jpeg":
		fileType = "image/jpeg"
	case ".heic":
		fileType = "image/heic"
	}

	timestamp, err := exifTimestamp(path)
	if err != nil {
		info, err := os.Stat(path)
		if err == nil {
			timestamp = info.ModTime()
		}
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	tags := parts[:len(parts)-1]

	return metadata{
		imageID:       hash,
		agentID:       agentID,
		fileName:      relPath,
		fileType:      fileType,
		fileTimestamp: timestamp,
		tags:          tags,
	}, nil
}

func exifTimestamp(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return time.Time{}, err
	}

	t, err := x.DateTime()
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func upload(path string, meta metadata) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	if err != nil {
		return err
	}

	writer.WriteField("image_id", meta.imageID)
	writer.WriteField("agent_id", meta.agentID)
	writer.WriteField("file_name", meta.fileName)
	writer.WriteField("file_type", meta.fileType)
	writer.WriteField("file_timestamp", meta.fileTimestamp.UTC().Format(time.RFC3339))
	writer.WriteField("tags", strings.Join(meta.tags, ","))

	writer.Close()

	resp, err := http.Post(
		os.Getenv("BACKEND_URL")+"/embed/image",
		writer.FormDataContentType(),
		&body,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Backend returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}