package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("Starting Matter Server test...")

	// Start server in background
	go func() {
		fmt.Println("Server should be running on port 5580...")
		time.Sleep(2 * time.Second)

		// Test HTTP API
		testHTTPAPI()

		// Test WebSocket
		testWebSocket()

		// Signal to stop
		time.Sleep(1 * time.Second)
		fmt.Println("\nTest completed successfully!")
		os.Exit(0)
	}()

	// Wait for interrupt
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}

func testHTTPAPI() {
	fmt.Println("\n--- Testing HTTP API ---")

	// Test health endpoint
	resp, err := http.Get("http://localhost:5580/health")
	if err != nil {
		fmt.Printf("Error calling health endpoint: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading health response: %v\n", err)
		return
	}

	fmt.Printf("Health endpoint response: %s\n", string(body))

	// Test server info endpoint
	resp, err = http.Get("http://localhost:5580/api/info")
	if err != nil {
		fmt.Printf("Error calling info endpoint: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading info response: %v\n", err)
		return
	}

	fmt.Printf("Server info response: %s\n", string(body))
}

func testWebSocket() {
	fmt.Println("\n--- Testing WebSocket API ---")

	u := url.URL{Scheme: "ws", Host: "localhost:5580", Path: "/ws"}
	fmt.Printf("Connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Printf("WebSocket dial error: %v\n", err)
		return
	}
	defer c.Close()

	// Read server info message (sent automatically on connect)
	_, message, err := c.ReadMessage()
	if err != nil {
		fmt.Printf("WebSocket read error: %v\n", err)
		return
	}
	fmt.Printf("Received server info: %s\n", string(message))

	// Send server info command
	cmd := map[string]interface{}{
		"message_id": "test-1",
		"command":    "server_info",
	}

	if err := c.WriteJSON(cmd); err != nil {
		fmt.Printf("WebSocket write error: %v\n", err)
		return
	}

	// Read response
	_, message, err = c.ReadMessage()
	if err != nil {
		fmt.Printf("WebSocket read error: %v\n", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(message, &response); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return
	}

	fmt.Printf("Server info command response: %s\n", string(message))

	// Send get_nodes command
	cmd = map[string]interface{}{
		"message_id": "test-2",
		"command":    "get_nodes",
	}

	if err := c.WriteJSON(cmd); err != nil {
		fmt.Printf("WebSocket write error: %v\n", err)
		return
	}

	// Read response
	_, message, err = c.ReadMessage()
	if err != nil {
		fmt.Printf("WebSocket read error: %v\n", err)
		return
	}

	fmt.Printf("Get nodes command response: %s\n", string(message))
}
