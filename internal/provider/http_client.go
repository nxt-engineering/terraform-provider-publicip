package provider

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"inet.af/netaddr"
)

func forceNetwork(client *http.Client, network string, sourceIP netaddr.IP) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
		// Mirrors http.DefaultTransport DialContext,
		// with the exception that 'network' and
		// eventually 'LocalAddr' are overwritten.
		// Based upon https://stackoverflow.com/a/69307638/172132

		log.Printf("Dial 🌐: Network: '%s' LocalAddr: '%s'", network, sourceIP.String())

		var dialer *net.Dialer
		if sourceIP.IsZero() {
			dialer = &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}
		} else {
			dialer = &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				LocalAddr: &net.TCPAddr{IP: net.ParseIP(sourceIP.String())},
			}
		}
		return dialer.DialContext(ctx, network, addr)
	}

	client.Transport = transport
}
