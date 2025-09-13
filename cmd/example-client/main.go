package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Message types matching the server's protocol
type CommandMessage struct {
	MessageID string                 `json:"message_id"`
	Command   string                 `json:"command"`
	Args      map[string]interface{} `json:"args,omitempty"`
}

type ResultMessage struct {
	MessageID string      `json:"message_id"`
	Result    interface{} `json:"result,omitempty"`
	ErrorCode int         `json:"error_code,omitempty"`
	Details   *string     `json:"details,omitempty"`
}

type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// mDNS discovery for finding matter-server
func discoverMatterServer(ctx context.Context, timeout time.Duration) (string, error) {
	fmt.Println("🔍 Discovering matter-server via mDNS...")

	// Listen on mDNS multicast address
	conn, err := net.ListenMulticastUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	})
	if err != nil {
		return "", fmt.Errorf("failed to listen on mDNS: %w", err)
	}
	defer conn.Close()

	// Query for matter-server.local
	query := buildDNSQuery("matter-server.local", 1) // A record

	// Send query
	_, err = conn.WriteToUDP(query, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	})
	if err != nil {
		return "", fmt.Errorf("failed to send mDNS query: %w", err)
	}

	fmt.Println("📡 Sent mDNS query for matter-server.local...")

	// Listen for responses with timeout
	conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 1500)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			continue
		}

		// Parse DNS response and look for A records
		if ip := parseDNSResponse(buf[:n], "matter-server.local"); ip != "" {
			fmt.Printf("✅ Found matter-server at %s (from %s)\n", ip, addr.IP)
			return ip, nil
		}
	}

	// Fallback: try localhost
	fmt.Println("⚠️ No mDNS response received, trying localhost...")
	return "127.0.0.1", nil
}

// Simplified DNS query builder
func buildDNSQuery(hostname string, recordType uint16) []byte {
	query := make([]byte, 0, 256)

	// DNS header
	query = append(query, 0x00, 0x00) // ID
	query = append(query, 0x01, 0x00) // Flags (standard query)
	query = append(query, 0x00, 0x01) // Questions
	query = append(query, 0x00, 0x00) // Answers
	query = append(query, 0x00, 0x00) // Authority RRs
	query = append(query, 0x00, 0x00) // Additional RRs

	// Question section
	parts := strings.Split(hostname, ".")
	for _, part := range parts {
		if part != "" {
			query = append(query, byte(len(part)))
			query = append(query, []byte(part)...)
		}
	}
	query = append(query, 0x00) // End of name

	// Query type and class
	query = append(query, byte(recordType>>8), byte(recordType)) // Type A
	query = append(query, 0x00, 0x01)                            // Class IN

	return query
}

// Simplified DNS response parser
func parseDNSResponse(buf []byte, hostname string) string {
	if len(buf) < 12 {
		return ""
	}

	// Check if it's a response
	if buf[2]&0x80 == 0 {
		return ""
	}

	answerCount := uint16(buf[6])<<8 | uint16(buf[7])
	if answerCount == 0 {
		return ""
	}

	// Skip header and questions to get to answers
	offset := 12

	// Skip questions
	questionCount := uint16(buf[4])<<8 | uint16(buf[5])
	for i := uint16(0); i < questionCount; i++ {
		// Skip name
		for offset < len(buf) && buf[offset] != 0 {
			if buf[offset]&0xc0 == 0xc0 {
				offset += 2
				break
			}
			offset += int(buf[offset]) + 1
		}
		if offset < len(buf) && buf[offset] == 0 {
			offset++
		}
		offset += 4 // Skip type and class
	}

	// Parse answers
	for i := uint16(0); i < answerCount && offset+10 < len(buf); i++ {
		// Skip name (could be compressed)
		if buf[offset]&0xc0 == 0xc0 {
			offset += 2
		} else {
			for offset < len(buf) && buf[offset] != 0 {
				offset += int(buf[offset]) + 1
			}
			if offset < len(buf) {
				offset++ // Skip null terminator
			}
		}

		if offset+10 > len(buf) {
			break
		}

		recordType := uint16(buf[offset])<<8 | uint16(buf[offset+1])
		dataLen := uint16(buf[offset+8])<<8 | uint16(buf[offset+9])
		offset += 10

		if recordType == 1 && dataLen == 4 && offset+4 <= len(buf) { // A record
			ip := net.IP(buf[offset : offset+4])
			return ip.String()
		}
		offset += int(dataLen)
	}

	return ""
}

// WebSocket client for communicating with matter-server
type MatterClient struct {
	conn   *websocket.Conn
	url    string
	logger func(string, ...interface{})
}

func NewMatterClient(serverIP string, port int) *MatterClient {
	return &MatterClient{
		url: fmt.Sprintf("ws://%s:%d/ws", serverIP, port),
		logger: func(format string, args ...interface{}) {
			log.Printf(format, args...)
		},
	}
}

func (mc *MatterClient) Connect(ctx context.Context) error {
	mc.logger("🔌 Connecting to matter-server at %s", mc.url)

	u, err := url.Parse(mc.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	mc.conn = conn
	mc.logger("✅ Connected to matter-server WebSocket")

	// Start message reader
	go mc.readMessages(ctx)

	return nil
}

func (mc *MatterClient) readMessages(ctx context.Context) {
	defer mc.conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg json.RawMessage
			err := mc.conn.ReadJSON(&msg)
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					mc.logger("❌ Error reading message: %v", err)
				}
				return
			}

			mc.handleMessage(msg)
		}
	}
}

func (mc *MatterClient) handleMessage(rawMsg json.RawMessage) {
	// Try to determine message type
	var msgType map[string]interface{}
	if err := json.Unmarshal(rawMsg, &msgType); err != nil {
		mc.logger("❌ Failed to parse message: %v", err)
		return
	}

	if _, hasResult := msgType["result"]; hasResult || msgType["error_code"] != nil {
		// Result message
		var result ResultMessage
		if err := json.Unmarshal(rawMsg, &result); err == nil {
			if result.ErrorCode != 0 {
				details := "unknown error"
				if result.Details != nil {
					details = *result.Details
				}
				mc.logger("❌ Command failed [%s]: error %d - %s", result.MessageID, result.ErrorCode, details)
			} else {
				mc.logger("✅ Command success [%s]: %v", result.MessageID, result.Result)
			}
		}
	} else if _, hasEvent := msgType["event"]; hasEvent {
		// Event message
		var eventMsg EventMessage
		if err := json.Unmarshal(rawMsg, &eventMsg); err == nil {
			mc.logger("📢 Event: %s - %v", eventMsg.Event, eventMsg.Data)
		}
	} else {
		mc.logger("📨 Raw message: %s", string(rawMsg))
	}
}

func (mc *MatterClient) SendCommand(command string, args map[string]interface{}) error {
	cmd := CommandMessage{
		MessageID: uuid.New().String(),
		Command:   command,
		Args:      args,
	}

	mc.logger("📤 Sending command: %s [%s]", command, cmd.MessageID)

	return mc.conn.WriteJSON(cmd)
}

func (mc *MatterClient) Close() error {
	if mc.conn != nil {
		return mc.conn.Close()
	}
	return nil
}

func main() {
	fmt.Println("🚀 Matter Server Example Client")
	fmt.Println("===============================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down...")
		cancel()
	}()

	// Discover matter-server via mDNS
	serverIP, err := discoverMatterServer(ctx, 5*time.Second)
	if err != nil {
		log.Fatalf("❌ Failed to discover matter-server: %v", err)
	}

	// Connect to matter-server
	client := NewMatterClient(serverIP, 5580)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("❌ Failed to connect to matter-server: %v", err)
	}
	defer client.Close()

	// Wait a moment for connection to stabilize
	time.Sleep(1 * time.Second)

	// Send example commands
	examples := []struct {
		name    string
		command string
		args    map[string]interface{}
	}{
		{
			name:    "Get Server Info",
			command: "server_info",
			args:    nil,
		},
		{
			name:    "Get Server Diagnostics",
			command: "diagnostics",
			args:    nil,
		},
		{
			name:    "Get All Nodes",
			command: "get_nodes",
			args:    nil,
		},
		{
			name:    "Start Listening",
			command: "start_listening",
			args:    nil,
		},
		{
			name:    "Discover Devices",
			command: "discover",
			args:    map[string]interface{}{},
		},
	}

	fmt.Println("\n📋 Sending example commands...")
	for i, example := range examples {
		fmt.Printf("\n%d. %s\n", i+1, example.name)
		if err := client.SendCommand(example.command, example.args); err != nil {
			log.Printf("❌ Failed to send %s command: %v", example.command, err)
		}
		time.Sleep(1 * time.Second) // Wait between commands
	}

	fmt.Println("\n⏳ Listening for responses and events for 10 seconds...")

	// Listen for responses and events
	select {
	case <-ctx.Done():
	case <-time.After(10 * time.Second):
		fmt.Println("⏰ Timeout reached")
	}

	fmt.Println("👋 Client shutting down...")
}
