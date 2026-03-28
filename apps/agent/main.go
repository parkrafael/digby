package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"agent/config"
	"agent/ledger"
	"agent/tunnel"
	"agent/uploader"
	"agent/watcher"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := config.Load()
	if err != nil && os.IsNotExist(err) {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Which folder should I watch? [%s]: ", filepath.Join(os.Getenv("HOME"), "Pictures"))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			input = filepath.Join(os.Getenv("HOME"), "Pictures")
		}
		cfg, err = config.Setup(input)
		if err != nil {
			log.Fatal("Failed to create config: ", err)
		}
		fmt.Printf("\nYour agent ID is: %s\n", cfg.AgentID)
		fmt.Println("Go to digbyapp.xyz and sign up with this ID.")
	} else if err != nil {
		log.Fatal("Config file is corrupted: ", err)
	}

	lgr, err := ledger.Load()
	if err != nil {
		log.Fatal("Failed to load ledger: ", err)
	}
	defer lgr.Close()

	pollUntilRegistered(cfg.AgentID)

	queue := make(chan string, 100)

	go watcher.Run(cfg.WatchFolder, queue, lgr)
	go uploader.Run(cfg.AgentID, cfg.WatchFolder, queue, lgr)
	go tunnel.Run(cfg.AgentID, cfg.WatchFolder)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	fmt.Println("\nShutting down...")
}

func pollUntilRegistered(agentID string) {
	fmt.Println("Waiting for you to register on digbyapp.xyz...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := http.Get(fmt.Sprintf("%s/agent/registered?agent_id=%s", os.Getenv("BACKEND_URL"), agentID))
		if err != nil {
			fmt.Println("Could not reach backend, retrying...")
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Starting daemon...")
			return
		}
	}
}
