package mdns

import (
	"net"
	"testing"

	"github.com/codefionn/go-matter-server/internal/logger"
)

func TestNewMatterZone(t *testing.T) {
	log := logger.NewConsoleLogger(logger.ErrorLevel)

	tests := []struct {
		name     string
		hostname string
		expected string
	}{
		{
			name:     "Empty hostname",
			hostname: "",
			expected: "matter-server.local",
		},
		{
			name:     "Hostname without .local",
			hostname: "test-server",
			expected: "test-server.local",
		},
		{
			name:     "Hostname with .local",
			hostname: "test-server.local",
			expected: "test-server.local",
		},
		{
			name:     "Complex hostname",
			hostname: "my-matter-device",
			expected: "my-matter-device.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zone := NewMatterZone(tt.hostname, log)
			if zone.GetHostname() != tt.expected {
				t.Errorf("Expected hostname %s, got %s", tt.expected, zone.GetHostname())
			}
		})
	}
}

func TestMatterZoneRecords(t *testing.T) {
	log := logger.NewConsoleLogger(logger.ErrorLevel)
	zone := NewMatterZone("test.local", log)

	// Mock some IP addresses for testing
	zone.ips = []net.IP{
		net.ParseIP("192.168.1.100"), // IPv4
		net.ParseIP("2001:db8::1"),   // IPv6
		net.ParseIP("10.0.0.50"),     // Another IPv4
	}

	tests := []struct {
		name       string
		question   Question
		expectA    int // Number of A records expected
		expectAAAA int // Number of AAAA records expected
	}{
		{
			name: "A record query for hostname",
			question: Question{
				Name:  "test.local",
				Type:  dnsTypeA,
				Class: 1,
			},
			expectA:    2, // Should return both IPv4 addresses
			expectAAAA: 0,
		},
		{
			name: "AAAA record query for hostname",
			question: Question{
				Name:  "test.local",
				Type:  dnsTypeAAAA,
				Class: 1,
			},
			expectA:    0,
			expectAAAA: 1, // Should return the IPv6 address
		},
		{
			name: "Query for different hostname",
			question: Question{
				Name:  "other.local",
				Type:  dnsTypeA,
				Class: 1,
			},
			expectA:    0, // Should not respond to different hostname
			expectAAAA: 0,
		},
		{
			name: "PTR query (unsupported)",
			question: Question{
				Name:  "test.local",
				Type:  dnsTypePTR,
				Class: 1,
			},
			expectA:    0, // Should not respond to PTR queries
			expectAAAA: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records := zone.Records(tt.question)

			aCount := 0
			aaaaCount := 0

			for _, r := range records {
				switch r.(type) {
				case *A:
					aCount++
				case *AAAA:
					aaaaCount++
				}
			}

			if aCount != tt.expectA {
				t.Errorf("Expected %d A records, got %d", tt.expectA, aCount)
			}
			if aaaaCount != tt.expectAAAA {
				t.Errorf("Expected %d AAAA records, got %d", tt.expectAAAA, aaaaCount)
			}
		})
	}
}

func TestMatterZoneCaseInsensitive(t *testing.T) {
	log := logger.NewConsoleLogger(logger.ErrorLevel)
	zone := NewMatterZone("Test-Server.local", log)

	// Mock an IPv4 address
	zone.ips = []net.IP{net.ParseIP("192.168.1.100")}

	// Test case-insensitive matching
	tests := []string{
		"test-server.local",
		"TEST-SERVER.LOCAL",
		"Test-Server.Local",
		"TeSt-SeRvEr.LoCaL",
	}

	for _, hostname := range tests {
		t.Run(hostname, func(t *testing.T) {
			question := Question{
				Name:  hostname,
				Type:  dnsTypeA,
				Class: 1,
			}

			records := zone.Records(question)
			if len(records) == 0 {
				t.Errorf("Expected records for hostname %s, got none", hostname)
			}
		})
	}
}

func TestUpdateIPs(t *testing.T) {
	log := logger.NewConsoleLogger(logger.ErrorLevel)
	zone := NewMatterZone("test.local", log)

	// Update IPs
	zone.UpdateIPs()

	// Should have some IPs from the system
	updatedCount := len(zone.GetIPs())
	if updatedCount == 0 {
		t.Error("Expected some IP addresses after update, got none")
	}

	// Test that it finds real network interfaces
	t.Logf("Found %d IP addresses on system", updatedCount)
}

func TestDNSTypeToString(t *testing.T) {
	tests := []struct {
		dnsType  uint16
		expected string
	}{
		{dnsTypeA, "A"},
		{dnsTypeAAAA, "AAAA"},
		{dnsTypePTR, "PTR"},
		{dnsTypeTXT, "TXT"},
		{dnsTypeSRV, "SRV"},
		{999, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := dnsTypeToString(tt.dnsType)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMatterZoneRecordContent(t *testing.T) {
	log := logger.NewConsoleLogger(logger.ErrorLevel)
	zone := NewMatterZone("test.local", log)

	testIPv4 := net.ParseIP("192.168.1.100")
	testIPv6 := net.ParseIP("2001:db8::1")
	zone.ips = []net.IP{testIPv4, testIPv6}

	// Test A record content
	question := Question{
		Name:  "test.local",
		Type:  dnsTypeA,
		Class: 1,
	}

	records := zone.Records(question)
	if len(records) != 1 {
		t.Fatalf("Expected 1 A record, got %d", len(records))
	}

	aRecord, ok := records[0].(*A)
	if !ok {
		t.Fatal("Expected A record type")
	}

	if !aRecord.A.Equal(testIPv4) {
		t.Errorf("Expected A record IP %s, got %s", testIPv4, aRecord.A)
	}

	if aRecord.Hdr.Name != "test.local" {
		t.Errorf("Expected A record name 'test.local', got %s", aRecord.Hdr.Name)
	}

	if aRecord.Hdr.Type != dnsTypeA {
		t.Errorf("Expected A record type %d, got %d", dnsTypeA, aRecord.Hdr.Type)
	}

	if aRecord.Hdr.TTL != 120 {
		t.Errorf("Expected A record TTL 120, got %d", aRecord.Hdr.TTL)
	}
}

func TestMatterZoneStringRepresentation(t *testing.T) {
	log := logger.NewConsoleLogger(logger.ErrorLevel)
	zone := NewMatterZone("test.local", log)

	testIPv4 := net.ParseIP("192.168.1.100")
	zone.ips = []net.IP{testIPv4}

	question := Question{
		Name:  "test.local",
		Type:  dnsTypeA,
		Class: 1,
	}

	records := zone.Records(question)
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	recordStr := records[0].String()
	expectedStr := "test.local\tA\t192.168.1.100"
	if recordStr != expectedStr {
		t.Errorf("Expected record string '%s', got '%s'", expectedStr, recordStr)
	}
}
