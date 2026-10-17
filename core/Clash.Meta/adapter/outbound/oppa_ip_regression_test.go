package outbound

import (
	"bytes"
	"context"
	stdtls "crypto/tls"
	"io"
	"net"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/transport/oppa"
)

// Exercise the actual outbound through a TLS server, not a duplicated encoder.
func TestOppaDirectIPAndDomainDestinations(t *testing.T) {
	for _, tc := range []struct{ name, host, ip, want string }{
		{"domain", "example.invalid", "192.0.2.1", "example.invalid"},
		{"ipv4", "", "149.154.167.50", "149.154.167.50"},
		{"ipv6", "", "2001:db8::42", "2001:db8::42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(nil)
			defer server.Close()
			port := server.Listener.Addr().(*net.TCPAddr).Port
			result := make(chan error, 1)
			// Stop the HTTP serving loop; use a separate listener with its generated cert.
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			port = listener.Addr().(*net.TCPAddr).Port
			go func() {
				raw, e := listener.Accept()
				if e != nil {
					result <- e
					return
				}
				defer raw.Close()
				_ = raw.SetDeadline(time.Now().Add(5 * time.Second))
				conn := stdtls.Server(raw, server.TLS)
				prefix := make([]byte, len("test-secret")+1)
				if _, e = io.ReadFull(conn, prefix); e != nil {
					result <- e
					return
				}
				if !bytes.Equal(prefix, append([]byte("test-secret"), oppa.CommandTCP)) {
					result <- io.ErrUnexpectedEOF
					return
				}
				address, e := oppa.ParseAddress(conn)
				if e == nil && (address.String() != net.JoinHostPort(tc.want, "443")) {
					e = io.ErrUnexpectedEOF
				}
				result <- e
			}()
			proxy, err := NewOppa(OppaOption{Name: "test", Server: "127.0.0.1", Port: port, Password: "test-secret", SkipCertVerify: true})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, err := proxy.DialContext(ctx, &C.Metadata{NetWork: C.TCP, Host: tc.host, DstIP: netip.MustParseAddr(tc.ip), DstPort: 443})
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if err = <-result; err != nil {
				t.Fatal(err)
			}
		})
	}
}
