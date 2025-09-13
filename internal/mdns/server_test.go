package mdns

import (
	"net"
	"testing"

	"github.com/codefionn/go-matter-server/internal/logger"
)

// MockZone implements the Zone interface for testing
type MockZone struct {
	records map[string][]Record
}

func NewMockZone() *MockZone {
	return &MockZone{
		records: make(map[string][]Record),
	}
}

func (mz *MockZone) Records(q Question) []Record {
	key := q.Name
	if records, exists := mz.records[key]; exists {
		var filtered []Record
		for _, record := range records {
			if record.Header().Type == q.Type || q.Type == 0 {
				filtered = append(filtered, record)
			}
		}
		return filtered
	}
	return nil
}

func (mz *MockZone) AddRecord(record Record) {
	name := record.Header().Name
	mz.records[name] = append(mz.records[name], record)
}

func TestNewServer(t *testing.T) {
	zone := NewMockZone()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	config := &Config{
		Zone:   zone,
		Logger: log,
	}

	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server.config != config {
		t.Error("Config not set correctly")
	}
	if server.logger != log {
		t.Error("Logger not set correctly")
	}
}

func TestNewServerWithoutZone(t *testing.T) {
	config := &Config{
		Logger: logger.NewConsoleLogger(logger.InfoLevel),
	}

	_, err := NewServer(config)
	if err == nil {
		t.Error("Expected error when creating server without zone")
	}
	if err.Error() != "zone is required" {
		t.Errorf("Expected 'zone is required' error, got: %v", err)
	}
}

func TestNewServerWithoutLogger(t *testing.T) {
	zone := NewMockZone()
	config := &Config{
		Zone: zone,
	}

	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server.logger == nil {
		t.Error("Logger should be created automatically")
	}
}

func TestRRHeader(t *testing.T) {
	header := &RR_Header{
		Name:   "test.local.",
		Type:   dnsTypeA,
		Class:  1,
		TTL:    300,
		Length: 4,
	}

	if header.Header() != header {
		t.Error("Header() should return itself")
	}
}

func TestARecord(t *testing.T) {
	ip := net.ParseIP("192.168.1.100")
	record := &A{
		Hdr: RR_Header{
			Name:  "test.local.",
			Type:  dnsTypeA,
			Class: 1,
			TTL:   300,
		},
		A: ip,
	}

	if record.Header() != &record.Hdr {
		t.Error("Header() should return pointer to Hdr")
	}

	expected := "test.local.\tA\t192.168.1.100"
	if record.String() != expected {
		t.Errorf("Expected %s, got %s", expected, record.String())
	}
}

func TestAAAARecord(t *testing.T) {
	ip := net.ParseIP("2001:db8::1")
	record := &AAAA{
		Hdr: RR_Header{
			Name:  "test.local.",
			Type:  dnsTypeAAAA,
			Class: 1,
			TTL:   300,
		},
		AAAA: ip,
	}

	if record.Header() != &record.Hdr {
		t.Error("Header() should return pointer to Hdr")
	}

	expected := "test.local.\tAAAA\t2001:db8::1"
	if record.String() != expected {
		t.Errorf("Expected %s, got %s", expected, record.String())
	}
}

func TestPTRRecord(t *testing.T) {
	record := &PTR{
		Hdr: RR_Header{
			Name:  "100.1.168.192.in-addr.arpa.",
			Type:  dnsTypePTR,
			Class: 1,
			TTL:   300,
		},
		Ptr: "test.local.",
	}

	if record.Header() != &record.Hdr {
		t.Error("Header() should return pointer to Hdr")
	}

	expected := "100.1.168.192.in-addr.arpa.\tPTR\ttest.local."
	if record.String() != expected {
		t.Errorf("Expected %s, got %s", expected, record.String())
	}
}

func TestTXTRecord(t *testing.T) {
	record := &TXT{
		Hdr: RR_Header{
			Name:  "test.local.",
			Type:  dnsTypeTXT,
			Class: 1,
			TTL:   300,
		},
		Txt: []string{"version=1.0", "path=/api"},
	}

	if record.Header() != &record.Hdr {
		t.Error("Header() should return pointer to Hdr")
	}

	expected := "test.local.\tTXT\t[version=1.0 path=/api]"
	if record.String() != expected {
		t.Errorf("Expected %s, got %s", expected, record.String())
	}
}

func TestSRVRecord(t *testing.T) {
	record := &SRV{
		Hdr: RR_Header{
			Name:  "_http._tcp.test.local.",
			Type:  dnsTypeSRV,
			Class: 1,
			TTL:   300,
		},
		Priority: 10,
		Weight:   20,
		Port:     80,
		Target:   "server.test.local.",
	}

	if record.Header() != &record.Hdr {
		t.Error("Header() should return pointer to Hdr")
	}

	expected := "_http._tcp.test.local.\tSRV\t10 20 80 server.test.local."
	if record.String() != expected {
		t.Errorf("Expected %s, got %s", expected, record.String())
	}
}

func TestMockZoneRecords(t *testing.T) {
	zone := NewMockZone()

	// Add some test records
	aRecord := &A{
		Hdr: RR_Header{Name: "test.local.", Type: dnsTypeA, Class: 1, TTL: 300},
		A:   net.ParseIP("192.168.1.100"),
	}

	ptrRecord := &PTR{
		Hdr: RR_Header{Name: "test.local.", Type: dnsTypePTR, Class: 1, TTL: 300},
		Ptr: "server.local.",
	}

	zone.AddRecord(aRecord)
	zone.AddRecord(ptrRecord)

	// Query for A records
	question := Question{
		Name:  "test.local.",
		Type:  dnsTypeA,
		Class: 1,
	}

	records := zone.Records(question)
	if len(records) != 1 {
		t.Errorf("Expected 1 A record, got %d", len(records))
	}
	if records[0].Header().Type != dnsTypeA {
		t.Errorf("Expected A record type %d, got %d", dnsTypeA, records[0].Header().Type)
	}

	// Query for PTR records
	question.Type = dnsTypePTR
	records = zone.Records(question)
	if len(records) != 1 {
		t.Errorf("Expected 1 PTR record, got %d", len(records))
	}

	// Query for all records (Type = 0)
	question.Type = 0
	records = zone.Records(question)
	if len(records) != 2 {
		t.Errorf("Expected 2 total records, got %d", len(records))
	}

	// Query for non-existent name
	question.Name = "nonexistent.local."
	records = zone.Records(question)
	if len(records) != 0 {
		t.Errorf("Expected 0 records for non-existent name, got %d", len(records))
	}
}

func TestServerShutdownWithoutStart(t *testing.T) {
	zone := NewMockZone()
	config := &Config{
		Zone:   zone,
		Logger: logger.NewConsoleLogger(logger.ErrorLevel),
	}

	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Should be able to shutdown without starting
	err = server.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestDNSConstants(t *testing.T) {
	expectedTypes := map[string]uint16{
		"A":    1,
		"AAAA": 28,
		"PTR":  12,
		"TXT":  16,
		"SRV":  33,
	}

	actualTypes := map[string]uint16{
		"A":    dnsTypeA,
		"AAAA": dnsTypeAAAA,
		"PTR":  dnsTypePTR,
		"TXT":  dnsTypeTXT,
		"SRV":  dnsTypeSRV,
	}

	for name, expected := range expectedTypes {
		if actual := actualTypes[name]; actual != expected {
			t.Errorf("Expected %s type %d, got %d", name, expected, actual)
		}
	}

	// Test multicast addresses
	if mdnsGroupIPv4 != "224.0.0.251" {
		t.Errorf("Expected IPv4 multicast address 224.0.0.251, got %s", mdnsGroupIPv4)
	}
	if mdnsGroupIPv6 != "ff02::fb" {
		t.Errorf("Expected IPv6 multicast address ff02::fb, got %s", mdnsGroupIPv6)
	}
	if mdnsPort != 5353 {
		t.Errorf("Expected mDNS port 5353, got %d", mdnsPort)
	}
}

func TestQuestion(t *testing.T) {
	question := Question{
		Name:  "test.local.",
		Type:  dnsTypeA,
		Class: 1,
	}

	if question.Name != "test.local." {
		t.Errorf("Expected name 'test.local.', got %s", question.Name)
	}
	if question.Type != dnsTypeA {
		t.Errorf("Expected type %d, got %d", dnsTypeA, question.Type)
	}
	if question.Class != 1 {
		t.Errorf("Expected class 1, got %d", question.Class)
	}
}

// Test interface compliance
func TestRecordInterfaceCompliance(t *testing.T) {
	var records []Record = []Record{
		&A{},
		&AAAA{},
		&PTR{},
		&TXT{},
		&SRV{},
	}

	for i, record := range records {
		if record.Header() == nil {
			t.Errorf("Record %d: Header() returned nil", i)
		}
		if record.String() == "" {
			t.Errorf("Record %d: String() returned empty string", i)
		}
	}
}

func TestServerInterfaceName(t *testing.T) {
	zone := NewMockZone()
	config := &Config{
		Zone:   zone,
		Logger: logger.NewConsoleLogger(logger.ErrorLevel),
	}

	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test with no interface (should return default)
	name := server.interfaceName()
	if name == "" {
		t.Error("interfaceName() should not return empty string")
	}

	// Test with specific interface
	iface, err := net.InterfaceByName("lo")
	if err == nil { // Only test if loopback interface exists
		config.Interface = iface
		server.config = config
		name = server.interfaceName()
		if name != "lo" {
			t.Errorf("Expected interface name 'lo', got %s", name)
		}
	}
}
