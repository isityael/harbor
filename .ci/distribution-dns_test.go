package acme

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"

	"github.com/miekg/dns"
)

func TestHarborDNSFallback(t *testing.T) {
	for _, test := range []struct {
		name      string
		truncated bool
		tcpUp     bool
		wantError bool
	}{
		{"normal-udp", false, true, false},
		{"truncated-udp-uses-tcp", true, true, false},
		{"tcp-unavailable-error", true, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			address := listener.Addr().String()
			packet, err := net.ListenPacket("udp", address)
			if err != nil {
				listener.Close()
				t.Fatal(err)
			}
			var udpHits, tcpHits atomic.Int32
			handler := func(tcp bool) dns.HandlerFunc {
				return func(writer dns.ResponseWriter, request *dns.Msg) {
					response := new(dns.Msg)
					response.SetReply(request)
					if tcp {
						tcpHits.Add(1)
					} else {
						udpHits.Add(1)
					}
					response.Truncated = !tcp && test.truncated
					if !response.Truncated {
						response.Answer = []dns.RR{&dns.TXT{Hdr: dns.RR_Header{Name: request.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60}, Txt: []string{"verified"}}}
					}
					if err := writer.WriteMsg(response); err != nil {
						t.Errorf("DNS response: %v", err)
					}
				}
			}
			udpServer := &dns.Server{PacketConn: packet, Handler: handler(false)}
			go udpServer.ActivateAndServe()
			t.Cleanup(func() { udpServer.Shutdown() })
			if test.tcpUp {
				tcpServer := &dns.Server{Listener: listener, Handler: handler(true)}
				go tcpServer.ActivateAndServe()
				t.Cleanup(func() { tcpServer.Shutdown() })
			} else {
				listener.Close()
			}
			response, err := dnsQuery("_acme.example.", dns.TypeTXT, []string{address}, true)
			if test.wantError {
				var networkError *net.OpError
				if err == nil || !errors.As(err, &networkError) || networkError.Net != "tcp" {
					t.Fatalf("TCP failure must propagate, got response=%v error=%v", response, err)
				}
			} else if err != nil || response == nil || response.Truncated || len(response.Answer) != 1 || response.Answer[0].(*dns.TXT).Txt[0] != "verified" {
				t.Fatalf("unexpected DNS response=%v error=%v", response, err)
			}
			if udpHits.Load() == 0 {
				t.Fatal("the UDP server was not queried")
			}
			wantTCP := int32(0)
			if test.truncated && test.tcpUp {
				wantTCP = 1
			}
			if tcpHits.Load() != wantTCP {
				t.Fatalf("TCP server hits=%d, want %d", tcpHits.Load(), wantTCP)
			}
		})
	}
	t.Run("other-network-error", func(t *testing.T) {
		if response, err := dnsQuery("_acme.example.", dns.TypeTXT, []string{"127.0.0.1:not-a-port"}, true); err == nil || response != nil {
			t.Fatalf("network failure must propagate without dereferencing a nil response: %v, %v", response, err)
		}
	})
}
