package provider

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"
)

func forceNetwork(client *http.Client, network string) {
	log.Printf("Force Network 🌐 %s", network)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
		log.Printf("Dial 🌐 %s", network)

		// Mirrors http.DefaultTransport DialContext,
		// with the exception that 'network' is overwritten.
		// Based upon https://stackoverflow.com/a/69307638/172132
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext(ctx, network, addr)
	}

	client.Transport = transport
}

func forceV4(client *http.Client) {
	forceNetwork(client, "tcp4")
}

func forceV6(client *http.Client) {
	forceNetwork(client, "tcp6")
}
