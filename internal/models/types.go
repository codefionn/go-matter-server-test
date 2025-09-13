package models

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents different types of events that can be emitted
type EventType string

const (
	EventTypeNodeAdded         EventType = "node_added"
	EventTypeNodeUpdated       EventType = "node_updated"
	EventTypeNodeRemoved       EventType = "node_removed"
	EventTypeNodeEvent         EventType = "node_event"
	EventTypeAttributeUpdated  EventType = "attribute_updated"
	EventTypeServerShutdown    EventType = "server_shutdown"
	EventTypeServerInfoUpdated EventType = "server_info_updated"
	EventTypeEndpointAdded     EventType = "endpoint_added"
	EventTypeEndpointRemoved   EventType = "endpoint_removed"
)

// APICommand represents different API commands available
type APICommand string

const (
	APICommandStartListening          APICommand = "start_listening"
	APICommandServerDiagnostics       APICommand = "diagnostics"
	APICommandServerInfo              APICommand = "server_info"
	APICommandGetNodes                APICommand = "get_nodes"
	APICommandGetNode                 APICommand = "get_node"
	APICommandCommissionWithCode      APICommand = "commission_with_code"
	APICommandCommissionOnNetwork     APICommand = "commission_on_network"
	APICommandSetWiFiCredentials      APICommand = "set_wifi_credentials"
	APICommandSetThreadDataset        APICommand = "set_thread_dataset"
	APICommandOpenCommissioningWindow APICommand = "open_commissioning_window"
	APICommandDiscover                APICommand = "discover"
	APICommandInterviewNode           APICommand = "interview_node"
	APICommandDeviceCommand           APICommand = "device_command"
	APICommandRemoveNode              APICommand = "remove_node"
	APICommandGetVendorNames          APICommand = "get_vendor_names"
	APICommandReadAttribute           APICommand = "read_attribute"
	APICommandWriteAttribute          APICommand = "write_attribute"
	APICommandPingNode                APICommand = "ping_node"
	APICommandGetNodeIPAddresses      APICommand = "get_node_ip_addresses"
	APICommandImportTestNode          APICommand = "import_test_node"
	APICommandCheckNodeUpdate         APICommand = "check_node_update"
	APICommandUpdateNode              APICommand = "update_node"
	APICommandSetDefaultFabricLabel   APICommand = "set_default_fabric_label"
	APICommandSetACLEntry             APICommand = "set_acl_entry"
	APICommandSetNodeBinding          APICommand = "set_node_binding"
)

// VendorInfo contains vendor information from CSA
type VendorInfo struct {
	VendorID             int    `json:"vendor_id"`
	VendorName           string `json:"vendor_name"`
	CompanyLegalName     string `json:"company_legal_name"`
	CompanyPreferredName string `json:"company_preferred_name"`
	VendorLandingPageURL string `json:"vendor_landing_page_url"`
	Creator              string `json:"creator"`
}

// MatterNodeData represents Matter node data as stored on the server
type MatterNodeData struct {
	NodeID                 int                     `json:"node_id"`
	DateCommissioned       time.Time               `json:"date_commissioned"`
	LastInterview          time.Time               `json:"last_interview"`
	InterviewVersion       int                     `json:"interview_version"`
	Available              bool                    `json:"available"`
	IsBridge               bool                    `json:"is_bridge"`
	Attributes             map[string]interface{}  `json:"attributes"`
	AttributeSubscriptions []AttributeSubscription `json:"attribute_subscriptions"`
}

// AttributeSubscription represents an attribute subscription
type AttributeSubscription struct {
	EndpointID  *int `json:"endpoint_id"`
	ClusterID   *int `json:"cluster_id"`
	AttributeID *int `json:"attribute_id"`
}

// MatterNodeEvent represents a Matter node event
type MatterNodeEvent struct {
	NodeID        int                    `json:"node_id"`
	EndpointID    int                    `json:"endpoint_id"`
	ClusterID     int                    `json:"cluster_id"`
	EventID       int                    `json:"event_id"`
	EventNumber   int                    `json:"event_number"`
	Priority      int                    `json:"priority"`
	Timestamp     int64                  `json:"timestamp"`
	TimestampType int                    `json:"timestamp_type"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// ServerDiagnostics contains full server dump for diagnostics
type ServerDiagnostics struct {
	Info   ServerInfoMessage `json:"info"`
	Nodes  []MatterNodeData  `json:"nodes"`
	Events []interface{}     `json:"events"`
}

// NodePingResult contains ping results for a node
type NodePingResult map[string]bool

// Message types for WebSocket communication

// CommandMessage represents a command from client to server or vice versa
type CommandMessage struct {
	MessageID string                 `json:"message_id"`
	Command   string                 `json:"command"`
	Args      map[string]interface{} `json:"args,omitempty"`
}

// ResultMessageBase is the base class for result messages
type ResultMessageBase struct {
	MessageID string `json:"message_id"`
}

// SuccessResultMessage is sent when a command executes successfully
type SuccessResultMessage struct {
	ResultMessageBase
	Result interface{} `json:"result"`
}

// ErrorResultMessage is sent when a command fails
type ErrorResultMessage struct {
	ResultMessageBase
	ErrorCode int     `json:"error_code"`
	Details   *string `json:"details,omitempty"`
}

// EventMessage is sent for stateless events
type EventMessage struct {
	Event EventType   `json:"event"`
	Data  interface{} `json:"data"`
}

// ServerInfoMessage contains server information sent to clients
type ServerInfoMessage struct {
	FabricID                  int    `json:"fabric_id"`
	CompressedFabricID        int64  `json:"compressed_fabric_id"`
	SchemaVersion             int    `json:"schema_version"`
	MinSupportedSchemaVersion int    `json:"min_supported_schema_version"`
	SDKVersion                string `json:"sdk_version"`
	WiFiCredentialsSet        bool   `json:"wifi_credentials_set"`
	ThreadCredentialsSet      bool   `json:"thread_credentials_set"`
	BluetoothEnabled          bool   `json:"bluetooth_enabled"`
}

// CommissionableNodeData represents a discovered commissionable node
type CommissionableNodeData struct {
	InstanceName           *string  `json:"instance_name,omitempty"`
	HostName               *string  `json:"host_name,omitempty"`
	Port                   *int     `json:"port,omitempty"`
	LongDiscriminator      *int     `json:"long_discriminator,omitempty"`
	VendorID               *int     `json:"vendor_id,omitempty"`
	ProductID              *int     `json:"product_id,omitempty"`
	CommissioningMode      *int     `json:"commissioning_mode,omitempty"`
	DeviceType             *int     `json:"device_type,omitempty"`
	DeviceName             *string  `json:"device_name,omitempty"`
	PairingInstruction     *string  `json:"pairing_instruction,omitempty"`
	PairingHint            *int     `json:"pairing_hint,omitempty"`
	MRPRetryIntervalIdle   *int     `json:"mrp_retry_interval_idle,omitempty"`
	MRPRetryIntervalActive *int     `json:"mrp_retry_interval_active,omitempty"`
	SupportsTCP            *bool    `json:"supports_tcp,omitempty"`
	Addresses              []string `json:"addresses,omitempty"`
	RotatingID             *string  `json:"rotating_id,omitempty"`
}

// CommissioningParameters contains commissioning parameters
type CommissioningParameters struct {
	SetupPinCode    int    `json:"setup_pin_code"`
	SetupManualCode string `json:"setup_manual_code"`
	SetupQRCode     string `json:"setup_qr_code"`
}

// UpdateSource represents sources for software updates
type UpdateSource string

const (
	UpdateSourceMainNetDCL UpdateSource = "main-net-dcl"
	UpdateSourceTestNetDCL UpdateSource = "test-net-dcl"
	UpdateSourceLocal      UpdateSource = "local"
)

// MatterSoftwareVersion represents Matter software version information
type MatterSoftwareVersion struct {
	VID                          int          `json:"vid"`
	PID                          int          `json:"pid"`
	SoftwareVersion              int          `json:"software_version"`
	SoftwareVersionString        string       `json:"software_version_string"`
	FirmwareInformation          *string      `json:"firmware_information,omitempty"`
	MinApplicableSoftwareVersion int          `json:"min_applicable_software_version"`
	MaxApplicableSoftwareVersion int          `json:"max_applicable_software_version"`
	ReleaseNotesURL              *string      `json:"release_notes_url,omitempty"`
	UpdateSource                 UpdateSource `json:"update_source"`
}

// EventCallback is a function type for event callbacks
type EventCallback func(eventType EventType, data interface{})

// GenerateMessageID generates a new message ID
func GenerateMessageID() string {
	return uuid.New().String()
}
