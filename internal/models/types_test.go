package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventType(t *testing.T) {
	tests := []struct {
		name     string
		event    EventType
		expected string
	}{
		{"NodeAdded", EventTypeNodeAdded, "node_added"},
		{"NodeUpdated", EventTypeNodeUpdated, "node_updated"},
		{"NodeRemoved", EventTypeNodeRemoved, "node_removed"},
		{"NodeEvent", EventTypeNodeEvent, "node_event"},
		{"AttributeUpdated", EventTypeAttributeUpdated, "attribute_updated"},
		{"ServerShutdown", EventTypeServerShutdown, "server_shutdown"},
		{"ServerInfoUpdated", EventTypeServerInfoUpdated, "server_info_updated"},
		{"EndpointAdded", EventTypeEndpointAdded, "endpoint_added"},
		{"EndpointRemoved", EventTypeEndpointRemoved, "endpoint_removed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.event) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.event))
			}
		})
	}
}

func TestAPICommand(t *testing.T) {
	tests := []struct {
		name     string
		command  APICommand
		expected string
	}{
		{"StartListening", APICommandStartListening, "start_listening"},
		{"ServerDiagnostics", APICommandServerDiagnostics, "diagnostics"},
		{"ServerInfo", APICommandServerInfo, "server_info"},
		{"GetNodes", APICommandGetNodes, "get_nodes"},
		{"GetNode", APICommandGetNode, "get_node"},
		{"CommissionWithCode", APICommandCommissionWithCode, "commission_with_code"},
		{"CommissionOnNetwork", APICommandCommissionOnNetwork, "commission_on_network"},
		{"SetWiFiCredentials", APICommandSetWiFiCredentials, "set_wifi_credentials"},
		{"SetThreadDataset", APICommandSetThreadDataset, "set_thread_dataset"},
		{"OpenCommissioningWindow", APICommandOpenCommissioningWindow, "open_commissioning_window"},
		{"Discover", APICommandDiscover, "discover"},
		{"InterviewNode", APICommandInterviewNode, "interview_node"},
		{"DeviceCommand", APICommandDeviceCommand, "device_command"},
		{"RemoveNode", APICommandRemoveNode, "remove_node"},
		{"GetVendorNames", APICommandGetVendorNames, "get_vendor_names"},
		{"ReadAttribute", APICommandReadAttribute, "read_attribute"},
		{"WriteAttribute", APICommandWriteAttribute, "write_attribute"},
		{"PingNode", APICommandPingNode, "ping_node"},
		{"GetNodeIPAddresses", APICommandGetNodeIPAddresses, "get_node_ip_addresses"},
		{"ImportTestNode", APICommandImportTestNode, "import_test_node"},
		{"CheckNodeUpdate", APICommandCheckNodeUpdate, "check_node_update"},
		{"UpdateNode", APICommandUpdateNode, "update_node"},
		{"SetDefaultFabricLabel", APICommandSetDefaultFabricLabel, "set_default_fabric_label"},
		{"SetACLEntry", APICommandSetACLEntry, "set_acl_entry"},
		{"SetNodeBinding", APICommandSetNodeBinding, "set_node_binding"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.command) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.command))
			}
		})
	}
}

func TestUpdateSource(t *testing.T) {
	tests := []struct {
		name     string
		source   UpdateSource
		expected string
	}{
		{"MainNetDCL", UpdateSourceMainNetDCL, "main-net-dcl"},
		{"TestNetDCL", UpdateSourceTestNetDCL, "test-net-dcl"},
		{"Local", UpdateSourceLocal, "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.source))
			}
		})
	}
}

func TestMatterNodeDataJSON(t *testing.T) {
	now := time.Now()
	node := MatterNodeData{
		NodeID:           123,
		DateCommissioned: now,
		LastInterview:    now,
		InterviewVersion: 1,
		Available:        true,
		IsBridge:         false,
		Attributes:       map[string]interface{}{"test": "value"},
		AttributeSubscriptions: []AttributeSubscription{
			{EndpointID: intPtr(1), ClusterID: intPtr(2), AttributeID: intPtr(3)},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal MatterNodeData: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled MatterNodeData
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal MatterNodeData: %v", err)
	}

	// Verify fields
	if unmarshaled.NodeID != node.NodeID {
		t.Errorf("Expected NodeID %d, got %d", node.NodeID, unmarshaled.NodeID)
	}
	if unmarshaled.Available != node.Available {
		t.Errorf("Expected Available %v, got %v", node.Available, unmarshaled.Available)
	}
	if unmarshaled.IsBridge != node.IsBridge {
		t.Errorf("Expected IsBridge %v, got %v", node.IsBridge, unmarshaled.IsBridge)
	}
	if len(unmarshaled.AttributeSubscriptions) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(unmarshaled.AttributeSubscriptions))
	}
}

func TestMatterNodeEventJSON(t *testing.T) {
	event := MatterNodeEvent{
		NodeID:        123,
		EndpointID:    1,
		ClusterID:     2,
		EventID:       3,
		EventNumber:   4,
		Priority:      5,
		Timestamp:     1234567890,
		TimestampType: 1,
		Data:          map[string]interface{}{"key": "value"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal MatterNodeEvent: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled MatterNodeEvent
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal MatterNodeEvent: %v", err)
	}

	// Verify fields
	if unmarshaled.NodeID != event.NodeID {
		t.Errorf("Expected NodeID %d, got %d", event.NodeID, unmarshaled.NodeID)
	}
	if unmarshaled.EndpointID != event.EndpointID {
		t.Errorf("Expected EndpointID %d, got %d", event.EndpointID, unmarshaled.EndpointID)
	}
	if unmarshaled.ClusterID != event.ClusterID {
		t.Errorf("Expected ClusterID %d, got %d", event.ClusterID, unmarshaled.ClusterID)
	}
}

func TestServerInfoMessageJSON(t *testing.T) {
	info := ServerInfoMessage{
		FabricID:                  1,
		CompressedFabricID:        123456789,
		SchemaVersion:             11,
		MinSupportedSchemaVersion: 1,
		SDKVersion:                "go-matter-server-1.0.0",
		WiFiCredentialsSet:        true,
		ThreadCredentialsSet:      false,
		BluetoothEnabled:          true,
	}

	// Test JSON marshaling
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal ServerInfoMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled ServerInfoMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ServerInfoMessage: %v", err)
	}

	// Verify fields
	if unmarshaled.FabricID != info.FabricID {
		t.Errorf("Expected FabricID %d, got %d", info.FabricID, unmarshaled.FabricID)
	}
	if unmarshaled.SchemaVersion != info.SchemaVersion {
		t.Errorf("Expected SchemaVersion %d, got %d", info.SchemaVersion, unmarshaled.SchemaVersion)
	}
	if unmarshaled.SDKVersion != info.SDKVersion {
		t.Errorf("Expected SDKVersion %s, got %s", info.SDKVersion, unmarshaled.SDKVersion)
	}
	if unmarshaled.BluetoothEnabled != info.BluetoothEnabled {
		t.Errorf("Expected BluetoothEnabled %v, got %v", info.BluetoothEnabled, unmarshaled.BluetoothEnabled)
	}
}

func TestCommandMessageJSON(t *testing.T) {
	cmd := CommandMessage{
		MessageID: "test-123",
		Command:   "server_info",
		Args: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("Failed to marshal CommandMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled CommandMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal CommandMessage: %v", err)
	}

	// Verify fields
	if unmarshaled.MessageID != cmd.MessageID {
		t.Errorf("Expected MessageID %s, got %s", cmd.MessageID, unmarshaled.MessageID)
	}
	if unmarshaled.Command != cmd.Command {
		t.Errorf("Expected Command %s, got %s", cmd.Command, unmarshaled.Command)
	}
	if len(unmarshaled.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(unmarshaled.Args))
	}
}

func TestSuccessResultMessageJSON(t *testing.T) {
	success := SuccessResultMessage{
		ResultMessageBase: ResultMessageBase{
			MessageID: "test-123",
		},
		Result: map[string]interface{}{
			"status": "ok",
			"data":   []string{"item1", "item2"},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(success)
	if err != nil {
		t.Fatalf("Failed to marshal SuccessResultMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SuccessResultMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SuccessResultMessage: %v", err)
	}

	// Verify fields
	if unmarshaled.MessageID != success.MessageID {
		t.Errorf("Expected MessageID %s, got %s", success.MessageID, unmarshaled.MessageID)
	}
}

func TestErrorResultMessageJSON(t *testing.T) {
	details := "Something went wrong"
	errorMsg := ErrorResultMessage{
		ResultMessageBase: ResultMessageBase{
			MessageID: "test-123",
		},
		ErrorCode: 500,
		Details:   &details,
	}

	// Test JSON marshaling
	data, err := json.Marshal(errorMsg)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResultMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled ErrorResultMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ErrorResultMessage: %v", err)
	}

	// Verify fields
	if unmarshaled.MessageID != errorMsg.MessageID {
		t.Errorf("Expected MessageID %s, got %s", errorMsg.MessageID, unmarshaled.MessageID)
	}
	if unmarshaled.ErrorCode != errorMsg.ErrorCode {
		t.Errorf("Expected ErrorCode %d, got %d", errorMsg.ErrorCode, unmarshaled.ErrorCode)
	}
	if unmarshaled.Details == nil || *unmarshaled.Details != details {
		t.Errorf("Expected Details %s, got %v", details, unmarshaled.Details)
	}
}

func TestEventMessageJSON(t *testing.T) {
	event := EventMessage{
		Event: EventTypeNodeAdded,
		Data: map[string]interface{}{
			"node_id": 123,
			"status":  "online",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal EventMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled EventMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal EventMessage: %v", err)
	}

	// Verify fields
	if unmarshaled.Event != event.Event {
		t.Errorf("Expected Event %s, got %s", event.Event, unmarshaled.Event)
	}
}

func TestMatterSoftwareVersion(t *testing.T) {
	version := MatterSoftwareVersion{
		VID:                          123,
		PID:                          456,
		SoftwareVersion:              789,
		SoftwareVersionString:        "1.0.0",
		FirmwareInformation:          stringPtr("Test firmware"),
		MinApplicableSoftwareVersion: 1,
		MaxApplicableSoftwareVersion: 999,
		ReleaseNotesURL:              stringPtr("https://example.com/release-notes"),
		UpdateSource:                 UpdateSourceMainNetDCL,
	}

	// Test JSON marshaling
	data, err := json.Marshal(version)
	if err != nil {
		t.Fatalf("Failed to marshal MatterSoftwareVersion: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled MatterSoftwareVersion
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal MatterSoftwareVersion: %v", err)
	}

	// Verify fields
	if unmarshaled.VID != version.VID {
		t.Errorf("Expected VID %d, got %d", version.VID, unmarshaled.VID)
	}
	if unmarshaled.SoftwareVersionString != version.SoftwareVersionString {
		t.Errorf("Expected SoftwareVersionString %s, got %s", version.SoftwareVersionString, unmarshaled.SoftwareVersionString)
	}
	if unmarshaled.UpdateSource != version.UpdateSource {
		t.Errorf("Expected UpdateSource %s, got %s", version.UpdateSource, unmarshaled.UpdateSource)
	}
}

func TestGenerateMessageID(t *testing.T) {
	id1 := GenerateMessageID()
	id2 := GenerateMessageID()

	// IDs should not be empty
	if id1 == "" {
		t.Error("Generated message ID should not be empty")
	}
	if id2 == "" {
		t.Error("Generated message ID should not be empty")
	}

	// IDs should be different
	if id1 == id2 {
		t.Error("Generated message IDs should be unique")
	}

	// IDs should be valid UUIDs (basic check)
	if len(id1) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(id1))
	}
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
