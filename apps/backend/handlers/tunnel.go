package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// store active tunnel connections mapped by agent_id
var (
	tunnels   = make(map[string]*websocket.Conn)
	tunnelsMu sync.Mutex
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type tunnelRequest struct {
	RequestID string `json:"request_id"`
	FileName  string `json:"file_name"`
}

type tunnelResponse struct {
	RequestID string `json:"request_id"`
	Bytes     []byte `json:"bytes"`
	Error     string `json:"error,omitempty"`
}

func AgentTunnel(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "missing agent_id", http.StatusBadRequest)
		return
	}

	// upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade failed:", err)
		return
	}
	defer conn.Close()

	// store connection
	tunnelsMu.Lock()
	tunnels[agentID] = conn
	tunnelsMu.Unlock()

	fmt.Printf("Agent %s connected\n", agentID)

	// block until connection drops
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	// remove connection on disconnect
	tunnelsMu.Lock()
	delete(tunnels, agentID)
	tunnelsMu.Unlock()

	fmt.Printf("Agent %s disconnected\n", agentID)
}

// SendTunnelRequest sends a request to the daemon and waits for the response
func SendTunnelRequest(agentID string, req tunnelRequest) (*tunnelResponse, error) {
	tunnelsMu.Lock()
	conn, ok := tunnels[agentID]
	tunnelsMu.Unlock()

	if !ok {
		return nil, fmt.Errorf("agent not connected")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, err
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	var resp tunnelResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}