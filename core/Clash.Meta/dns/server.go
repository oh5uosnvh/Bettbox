package dns

import (
	"context"
	"net"
	"sync"

	"github.com/metacubex/mihomo/common/sockopt"
	"github.com/metacubex/mihomo/component/resolver"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/log"

	D "github.com/miekg/dns"
)

var (
	recreateMu sync.Mutex
	address    string
	server     = &Server{}

	dnsDefaultTTL uint32 = 600
)

type Server struct {
	service   resolver.Service
	tcpServer *D.Server
	udpServer *D.Server
}

type serverHandler struct {
	*Server
	isUDP bool
}

// ServeDNS implement D.Handler ServeDNS
func (s serverHandler) ServeDNS(w D.ResponseWriter, r *D.Msg) {
	msg, err := s.service.ServeMsg(context.Background(), r)
	if err != nil {
		m := new(D.Msg)
		m.SetRcode(r, D.RcodeServerFailure)
		// does not matter if this write fails
		w.WriteMsg(m)
		return
	}
	if s.isUDP {
		// RFC 6891: fit the reply into the client's advertised buffer size,
		// setting the TC bit if records must be dropped; 512 when no OPT present
		msg.Truncate(resolver.RequestUDPSize(r))
	}
	msg.Compress = true
	w.WriteMsg(msg)
}

func (s *Server) UDPHandler() D.Handler {
	return serverHandler{Server: s, isUDP: true}
}

func (s *Server) TCPHandler() D.Handler {
	return serverHandler{Server: s, isUDP: false}
}

func (s *Server) SetService(service resolver.Service) {
	s.service = service
}

func ReCreateServer(addr string, lc C.InboundListenConfig, service resolver.Service) {
	recreateMu.Lock()
	defer recreateMu.Unlock()
	if addr == address && service != nil {
		server.SetService(service)
		return
	}

	if server.tcpServer != nil {
		_ = server.tcpServer.Shutdown()
		server.tcpServer = nil
	}

	if server.udpServer != nil {
		_ = server.udpServer.Shutdown()
		server.udpServer = nil
	}

	server.service = nil
	address = ""

	if addr == "" || lc == nil || service == nil {
		return
	}

	var err error
	defer func() {
		if err != nil {
			log.Errorln("Start DNS server error: %s", err.Error())
		}
	}()

	_, port, err := net.SplitHostPort(addr)
	if port == "0" || port == "" || err != nil {
		return
	}

	address = addr
	server = &Server{service: service}

	// Bind before publishing the replacement so Shutdown never races a pending bind.
	current := server
	p, udpErr := lc.ListenPacket(context.Background(), "udp", addr)
	if udpErr != nil {
		log.Errorln("Start DNS server(UDP) error: %s", udpErr)
	} else {
		if err := sockopt.UDPReuseaddr(p); err != nil {
			log.Warnln("Failed to Reuse UDP Address: %s", err)
		}
		current.udpServer = &D.Server{Addr: addr, PacketConn: p, Handler: current.UDPHandler()}
	}
	l, tcpErr := lc.Listen(context.Background(), "tcp", addr)
	if tcpErr != nil {
		log.Errorln("Start DNS server(TCP) error: %s", tcpErr)
	} else {
		current.tcpServer = &D.Server{Addr: addr, Listener: l, Handler: current.TCPHandler()}
	}
	if udpErr != nil || tcpErr != nil {
		if p != nil {
			_ = p.Close()
		}
		if l != nil {
			_ = l.Close()
		}
		current.udpServer = nil
		current.tcpServer = nil
		address = ""
		return
	}
	udpReady, tcpReady := make(chan struct{}), make(chan struct{})
	udpServer, tcpServer := current.udpServer, current.tcpServer
	udpServer.NotifyStartedFunc = func() { close(udpReady) }
	tcpServer.NotifyStartedFunc = func() { close(tcpReady) }
	go func() { _ = udpServer.ActivateAndServe() }()
	go func() { _ = tcpServer.ActivateAndServe() }()
	<-udpReady
	<-tcpReady
	log.Infoln("DNS server listening at: %s", addr)
}
