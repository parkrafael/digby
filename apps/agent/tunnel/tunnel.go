package tunnel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type request struct {
	RequestID string `json:"request_id"`
	FileName  string `json:"file_name"`
}

type response struct {
	RequestID string `json:"request_id"`
	Bytes     []byte `json:"bytes"`
	Error     string `json:"error,omitempty"`
}

func Run(agentID string, watchFolder string) {
	for {
		err := connect(agentID, watchFolder)
		if err != nil {
			fmt.Println("Tunnel disconnected:", err)
		}
		fmt.Println("Reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

func connect(agentID string, watchFolder string) error {
	wsURL := strings.Replace(os.Getenv("BACKEND_URL"), "http://", "ws://", 1)
	url := fmt.Sprintf("%s/agent/tunnel?agent_id=%s", wsURL, agentID)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("Dial failed: %w", err)
	}
	defer conn.Close()

	fmt.Println("Tunnel connected.")
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("Read failed: %w", err)
		}

		var req request
		err = json.Unmarshal(msg, &req)
		if err != nil {
			fmt.Println("Could not parse message:", err)
			continue
		}

		go handleRequest(conn, req, watchFolder)
	}
}

func handleRequest(conn *websocket.Conn, req request, watchFolder string) {
	fullPath := filepath.Join(watchFolder, req.FileName)
	bytes, err := os.ReadFile(fullPath)

	var resp response
	resp.RequestID = req.RequestID
	if err != nil {
		resp.Error = fmt.Sprintf("Could not read file: %v", err)
	} else {
		resp.Bytes = bytes
	}

	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Could not marshal response:", err)
		return
	}

	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		fmt.Println("Could not write to tunnel:", err)
	}
}