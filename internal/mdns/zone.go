package mdns

import (
	"net"
	"strings"

	"github.com/codefionn/go-matter-server/internal/logger"
)

// MatterZone implements a DNS zone for Matter server hostname advertisement
type MatterZone struct {
	hostname string
	logger   *logger.Logger
	ips      []net.IP
}

// NewMatterZone creates a new mDNS zone for the Matter server
func NewMatterZone(hostname string, log *logger.Logger) *MatterZone {
	if hostname == "" {
		hostname = "matter-server"
	}

	// Ensure hostname ends with .local
	if !strings.HasSuffix(hostname, ".local") {
		hostname = hostname + ".local"
	}

	zone := &MatterZone{
		hostname: hostname,
		logger:   log,
	}

	// Get local IP addresses
	zone.updateIPs()

	return zone
}

// Records implements the Zone interface
func (z *MatterZone) Records(q Question) []Record {
	// Normalize query name
	qname := strings.ToLower(q.Name)
	hostname := strings.ToLower(z.hostname)

	z.logger.Debug("mDNS query",
		logger.String("question", qname),
		logger.String("type", dnsTypeToString(q.Type)),
		logger.String("hostname", hostname),
	)

	var records []Record

	// Only respond to queries for our hostname
	if qname != hostname {
		return records
	}

	switch q.Type {
	case dnsTypeA:
		// Return IPv4 addresses
		for _, ip := range z.ips {
			if ip.To4() != nil {
				records = append(records, &A{
					Hdr: RR_Header{
						Name:  z.hostname,
						Type:  dnsTypeA,
						Class: 1, // IN
						TTL:   120,
					},
					A: ip,
				})
			}
		}
	case dnsTypeAAAA:
		// Return IPv6 addresses
		for _, ip := range z.ips {
			if ip.To4() == nil && !ip.IsLoopback() {
				records = append(records, &AAAA{
					Hdr: RR_Header{
						Name:  z.hostname,
						Type:  dnsTypeAAAA,
						Class: 1, // IN
						TTL:   120,
					},
					AAAA: ip,
				})
			}
		}
	}

	z.logger.Debug("mDNS response",
		logger.String("hostname", hostname),
		logger.Int("records", len(records)),
	)

	return records
}

// UpdateIPs refreshes the list of local IP addresses
func (z *MatterZone) UpdateIPs() {
	z.updateIPs()
}

func (z *MatterZone) updateIPs() {
	var ips []net.IP

	interfaces, err := net.Interfaces()
	if err != nil {
		z.logger.Error("Failed to get network interfaces", logger.ErrorField(err))
		return
	}

	for _, iface := range interfaces {
		// Skip down interfaces and loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback and link-local addresses
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			ips = append(ips, ip)
		}
	}

	z.ips = ips
	z.logger.Debug("Updated mDNS IP addresses", logger.Int("count", len(ips)))
}

// GetHostname returns the advertised hostname
func (z *MatterZone) GetHostname() string {
	return z.hostname
}

// GetIPs returns the current list of IP addresses
func (z *MatterZone) GetIPs() []net.IP {
	return z.ips
}

func dnsTypeToString(dnsType uint16) string {
	switch dnsType {
	case dnsTypeA:
		return "A"
	case dnsTypeAAAA:
		return "AAAA"
	case dnsTypePTR:
		return "PTR"
	case dnsTypeTXT:
		return "TXT"
	case dnsTypeSRV:
		return "SRV"
	default:
		return "UNKNOWN"
	}
}
