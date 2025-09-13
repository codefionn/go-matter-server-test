package mdns

import (
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/codefionn/go-matter-server/internal/logger"
)

const (
	// mDNS multicast addresses
	mdnsGroupIPv4 = "224.0.0.251"
	mdnsGroupIPv6 = "ff02::fb"
	mdnsPort      = 5353

	// DNS constants
	dnsTypeA    = 1
	dnsTypeAAAA = 28
	dnsTypePTR  = 12
	dnsTypeTXT  = 16
	dnsTypeSRV  = 33
)

// Config holds the configuration for the mDNS server
type Config struct {
	Interface *net.Interface
	Logger    *logger.Logger
	Zone      Zone
}

// Zone defines the DNS records that the server will respond to
type Zone interface {
	Records(q Question) []Record
}

// Question represents a DNS query
type Question struct {
	Name  string
	Type  uint16
	Class uint16
}

// Record represents a DNS resource record
type Record interface {
	Header() *RR_Header
	String() string
}

// RR_Header is the header of a DNS resource record
type RR_Header struct {
	Name   string
	Type   uint16
	Class  uint16
	TTL    uint32
	Length uint16
}

func (h *RR_Header) Header() *RR_Header {
	return h
}

// A record for IPv4 addresses
type A struct {
	Hdr RR_Header
	A   net.IP
}

func (r *A) Header() *RR_Header { return &r.Hdr }
func (r *A) String() string     { return fmt.Sprintf("%s\tA\t%s", r.Hdr.Name, r.A) }

// AAAA record for IPv6 addresses
type AAAA struct {
	Hdr  RR_Header
	AAAA net.IP
}

func (r *AAAA) Header() *RR_Header { return &r.Hdr }
func (r *AAAA) String() string     { return fmt.Sprintf("%s\tAAAA\t%s", r.Hdr.Name, r.AAAA) }

// PTR record for pointer queries
type PTR struct {
	Hdr RR_Header
	Ptr string
}

func (r *PTR) Header() *RR_Header { return &r.Hdr }
func (r *PTR) String() string     { return fmt.Sprintf("%s\tPTR\t%s", r.Hdr.Name, r.Ptr) }

// TXT record for text data
type TXT struct {
	Hdr RR_Header
	Txt []string
}

func (r *TXT) Header() *RR_Header { return &r.Hdr }
func (r *TXT) String() string     { return fmt.Sprintf("%s\tTXT\t%v", r.Hdr.Name, r.Txt) }

// SRV record for service location
type SRV struct {
	Hdr      RR_Header
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string
}

func (r *SRV) Header() *RR_Header { return &r.Hdr }
func (r *SRV) String() string {
	return fmt.Sprintf("%s\tSRV\t%d %d %d %s", r.Hdr.Name, r.Priority, r.Weight, r.Port, r.Target)
}

// Server represents an mDNS server
type Server struct {
	config   *Config
	shutdown atomic.Bool
	ipv4conn *net.UDPConn
	ipv6conn *net.UDPConn
	logger   *logger.Logger
}

// NewServer creates a new mDNS server
func NewServer(config *Config) (*Server, error) {
	if config.Zone == nil {
		return nil, fmt.Errorf("zone is required")
	}

	if config.Logger == nil {
		config.Logger = logger.NewConsoleLogger(logger.InfoLevel)
	}

	return &Server{
		config: config,
		logger: config.Logger,
	}, nil
}

// Start begins listening for mDNS queries
func (s *Server) Start() error {
	var err error

	// Setup IPv4 listener
	if s.ipv4conn, err = s.setupIPv4(); err != nil {
		return fmt.Errorf("failed to setup IPv4: %w", err)
	}

	// Setup IPv6 listener
	if s.ipv6conn, err = s.setupIPv6(); err != nil {
		s.ipv4conn.Close()
		return fmt.Errorf("failed to setup IPv6: %w", err)
	}

	// Start receiving goroutines
	go s.recv(s.ipv4conn, false)
	go s.recv(s.ipv6conn, true)

	s.logger.Info("mDNS server started", logger.String("interface", s.interfaceName()))
	return nil
}

// Shutdown stops the mDNS server
func (s *Server) Shutdown() error {
	s.shutdown.Store(true)

	var errs []error

	if s.ipv4conn != nil {
		if err := s.ipv4conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.ipv6conn != nil {
		if err := s.ipv6conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	s.logger.Info("mDNS server shutdown")

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

func (s *Server) setupIPv4() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", mdnsGroupIPv4, mdnsPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenMulticastUDP("udp4", s.config.Interface, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *Server) setupIPv6() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp6", fmt.Sprintf("[%s]:%d", mdnsGroupIPv6, mdnsPort))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenMulticastUDP("udp6", s.config.Interface, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *Server) recv(conn *net.UDPConn, ipv6 bool) {
	buf := make([]byte, 65536)

	for !s.shutdown.Load() {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if !s.shutdown.Load() {
				s.logger.Error("Failed to read UDP packet", logger.ErrorField(err))
			}
			continue
		}

		if err := s.parsePacket(buf[:n], from, conn, ipv6); err != nil {
			s.logger.Debug("Failed to parse packet", logger.ErrorField(err))
		}
	}
}

func (s *Server) parsePacket(buf []byte, from *net.UDPAddr, conn *net.UDPConn, ipv6 bool) error {
	msg, err := parseDNSMessage(buf)
	if err != nil {
		return err
	}

	// Only handle queries
	if msg.Response {
		return nil
	}

	// Must be standard query
	if msg.Opcode != 0 {
		return nil
	}

	// Must not have response code set
	if msg.Rcode != 0 {
		return nil
	}

	return s.handleQuery(msg, from, conn, ipv6)
}

func (s *Server) handleQuery(msg *dnsMessage, from *net.UDPAddr, conn *net.UDPConn, ipv6 bool) error {
	if len(msg.Questions) == 0 {
		return nil
	}

	response := &dnsMessage{
		ID:                 msg.ID,
		Response:           true,
		Opcode:             0,
		Authoritative:      true,
		Truncated:          false,
		RecursionDesired:   false,
		RecursionAvailable: false,
		Rcode:              0,
	}

	for _, q := range msg.Questions {
		question := Question{
			Name:  q.Name,
			Type:  q.Type,
			Class: q.Class,
		}

		records := s.config.Zone.Records(question)
		for _, r := range records {
			response.Answers = append(response.Answers, dnsRecord{
				Name:  r.Header().Name,
				Type:  r.Header().Type,
				Class: r.Header().Class,
				TTL:   r.Header().TTL,
				Data:  s.encodeRecordData(r),
			})
		}
	}

	if len(response.Answers) > 0 {
		return s.sendResponse(response, from, conn)
	}

	return nil
}

func (s *Server) sendResponse(msg *dnsMessage, to *net.UDPAddr, conn *net.UDPConn) error {
	buf, err := encodeDNSMessage(msg)
	if err != nil {
		return err
	}

	_, err = conn.WriteToUDP(buf, to)
	if err != nil {
		s.logger.Error("Failed to send response", logger.ErrorField(err))
	}

	return err
}

func (s *Server) encodeRecordData(r Record) []byte {
	switch rec := r.(type) {
	case *A:
		return rec.A.To4()
	case *AAAA:
		return rec.AAAA.To16()
	case *PTR:
		return encodeName(rec.Ptr)
	case *TXT:
		return encodeTXT(rec.Txt)
	case *SRV:
		return encodeSRV(rec.Priority, rec.Weight, rec.Port, rec.Target)
	default:
		return nil
	}
}

func (s *Server) interfaceName() string {
	if s.config.Interface == nil {
		return "all"
	}
	return s.config.Interface.Name
}

// Simple DNS message structure for parsing
type dnsMessage struct {
	ID                 uint16
	Response           bool
	Opcode             uint8
	Authoritative      bool
	Truncated          bool
	RecursionDesired   bool
	RecursionAvailable bool
	Rcode              uint8
	Questions          []dnsQuestion
	Answers            []dnsRecord
}

type dnsQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

type dnsRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  []byte
}

// Simplified DNS message parsing and encoding
func parseDNSMessage(buf []byte) (*dnsMessage, error) {
	if len(buf) < 12 {
		return nil, fmt.Errorf("DNS message too short")
	}

	msg := &dnsMessage{
		ID:                 uint16(buf[0])<<8 | uint16(buf[1]),
		Response:           buf[2]&0x80 != 0,
		Opcode:             (buf[2] >> 3) & 0x0f,
		Authoritative:      buf[2]&0x04 != 0,
		Truncated:          buf[2]&0x02 != 0,
		RecursionDesired:   buf[2]&0x01 != 0,
		RecursionAvailable: buf[3]&0x80 != 0,
		Rcode:              buf[3] & 0x0f,
	}

	qdCount := uint16(buf[4])<<8 | uint16(buf[5])

	offset := 12
	for i := uint16(0); i < qdCount; i++ {
		name, newOffset, err := parseName(buf, offset)
		if err != nil {
			return nil, err
		}

		if newOffset+4 > len(buf) {
			return nil, fmt.Errorf("question truncated")
		}

		q := dnsQuestion{
			Name:  name,
			Type:  uint16(buf[newOffset])<<8 | uint16(buf[newOffset+1]),
			Class: uint16(buf[newOffset+2])<<8 | uint16(buf[newOffset+3]),
		}

		msg.Questions = append(msg.Questions, q)
		offset = newOffset + 4
	}

	return msg, nil
}

func encodeDNSMessage(msg *dnsMessage) ([]byte, error) {
	buf := make([]byte, 12)

	buf[0] = byte(msg.ID >> 8)
	buf[1] = byte(msg.ID)

	if msg.Response {
		buf[2] |= 0x80
	}
	buf[2] |= (msg.Opcode & 0x0f) << 3
	if msg.Authoritative {
		buf[2] |= 0x04
	}
	if msg.Truncated {
		buf[2] |= 0x02
	}
	if msg.RecursionDesired {
		buf[2] |= 0x01
	}
	if msg.RecursionAvailable {
		buf[3] |= 0x80
	}
	buf[3] |= msg.Rcode & 0x0f

	buf[4] = byte(len(msg.Questions) >> 8)
	buf[5] = byte(len(msg.Questions))
	buf[6] = byte(len(msg.Answers) >> 8)
	buf[7] = byte(len(msg.Answers))

	for _, q := range msg.Questions {
		nameBytes := encodeName(q.Name)
		buf = append(buf, nameBytes...)
		buf = append(buf, byte(q.Type>>8), byte(q.Type))
		buf = append(buf, byte(q.Class>>8), byte(q.Class))
	}

	for _, r := range msg.Answers {
		nameBytes := encodeName(r.Name)
		buf = append(buf, nameBytes...)
		buf = append(buf, byte(r.Type>>8), byte(r.Type))
		buf = append(buf, byte(r.Class>>8), byte(r.Class))
		buf = append(buf, byte(r.TTL>>24), byte(r.TTL>>16), byte(r.TTL>>8), byte(r.TTL))
		buf = append(buf, byte(len(r.Data)>>8), byte(len(r.Data)))
		buf = append(buf, r.Data...)
	}

	return buf, nil
}

func parseName(buf []byte, offset int) (string, int, error) {
	var name []string
	original := offset
	jumped := false

	for offset < len(buf) {
		length := int(buf[offset])
		if length == 0 {
			offset++
			break
		}

		if length&0xc0 == 0xc0 {
			if !jumped {
				original = offset + 2
			}
			offset = int(buf[offset]&0x3f)<<8 | int(buf[offset+1])
			jumped = true
			continue
		}

		if offset+1+length >= len(buf) {
			return "", 0, fmt.Errorf("name extends past buffer")
		}

		name = append(name, string(buf[offset+1:offset+1+length]))
		offset += 1 + length
	}

	if !jumped {
		original = offset
	}

	return strings.Join(name, "."), original, nil
}

func encodeName(name string) []byte {
	if name == "." {
		return []byte{0}
	}

	parts := strings.Split(name, ".")
	var buf []byte

	for _, part := range parts {
		if part != "" {
			buf = append(buf, byte(len(part)))
			buf = append(buf, []byte(part)...)
		}
	}

	buf = append(buf, 0)
	return buf
}

func encodeTXT(txt []string) []byte {
	var buf []byte
	for _, t := range txt {
		if len(t) > 255 {
			t = t[:255]
		}
		buf = append(buf, byte(len(t)))
		buf = append(buf, []byte(t)...)
	}
	return buf
}

func encodeSRV(priority, weight, port uint16, target string) []byte {
	buf := make([]byte, 6)
	buf[0] = byte(priority >> 8)
	buf[1] = byte(priority)
	buf[2] = byte(weight >> 8)
	buf[3] = byte(weight)
	buf[4] = byte(port >> 8)
	buf[5] = byte(port)

	targetBytes := encodeName(target)
	return append(buf, targetBytes...)
}
