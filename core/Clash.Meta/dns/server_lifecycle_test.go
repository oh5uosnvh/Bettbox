package dns

import (
	"context"
	"net"
	"testing"

	D "github.com/miekg/dns"
)

type testListen struct{ net.ListenConfig }
type testService struct{}

func (testService) ServeMsg(_ context.Context, m *D.Msg) (*D.Msg, error) {
	return new(D.Msg).SetReply(m), nil
}

func TestRecreateBindsAndClosesBeforeNextGeneration(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := probe.Addr().String()
	probe.Close()
	defer ReCreateServer("", nil, nil)
	for i := 0; i < 40; i++ {
		ReCreateServer(addr, &testListen{}, testService{})
		if server.tcpServer == nil || server.udpServer == nil {
			t.Fatal("listener missing")
		}
		ReCreateServer("", nil, nil)
		tcp, e := net.Listen("tcp", addr)
		if e != nil {
			t.Fatal(e)
		}
		tcp.Close()
		udp, e := net.ListenPacket("udp", addr)
		if e != nil {
			t.Fatal(e)
		}
		udp.Close()
	}
}
