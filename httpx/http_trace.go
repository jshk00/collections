package httpx

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http/httptrace"
	"time"
)

type TraceInfo struct {
	// DNSLookup is the duration that transport took to perform
	// DNS lookup.
	DNSLookup time.Duration `json:"dns_lookup_time"`
	// ConnTime is the duration it took to obtain a successful connection.
	ConnTime time.Duration `json:"connection_time"`
	// TCPConnTime is the duration it took to obtain the TCP connection.
	TCPConnTime time.Duration `json:"tcp_connection_time"`
	// TLSHandshake is the duration of the TLS handshake.
	TLSHandshake time.Duration `json:"tls_handshake_time"`
	// ServerTime is the server's duration for responding to the first byte.
	ServerTime time.Duration `json:"server_time"`
	// ResponseTime is the duration since the first response byte from the server to
	// request completion.
	ResponseTime time.Duration `json:"response_time"`
	// TotalTime is the duration of the total time request taken end-to-end.
	TotalTime time.Duration `json:"total_time"`
	// IsConnReused is whether this connection has been previously
	// used for another HTTP request.
	IsConnReused bool `json:"is_connection_reused"`
	// IsConnWasIdle is whether this connection was obtained from an
	// idle pool.
	IsConnWasIdle bool `json:"is_connection_was_idle"`
	// ConnIdleTime is the duration how long the connection that was previously
	// idle, if IsConnWasIdle is true.
	ConnIdleTime time.Duration `json:"connection_idle_time"`
	// RequestAttempt is to represent the request attempt made during
	// request execution flow, including retry count.
	RequestAttempt int `json:"request_attempt"`
	// RemoteAddr returns the remote network address.
	RemoteAddr string `json:"remote_address"`
}

// String method returns string representation of request trace information.
func (ti TraceInfo) String() string {
	return fmt.Sprintf(`TRACE INFO:
  DNSLookupTime : %v
  ConnTime      : %v
  TCPConnTime   : %v
  TLSHandshake  : %v
  ServerTime    : %v
  ResponseTime  : %v
  TotalTime     : %v
  IsConnReused  : %v
  IsConnWasIdle : %v
  ConnIdleTime  : %v
  RequestAttempt: %v
  RemoteAddr    : %v`, ti.DNSLookup, ti.ConnTime, ti.TCPConnTime,
		ti.TLSHandshake, ti.ServerTime, ti.ResponseTime, ti.TotalTime,
		ti.IsConnReused, ti.IsConnWasIdle, ti.ConnIdleTime, ti.RequestAttempt,
		ti.RemoteAddr)
}

func getTraceLogPrinting(ctx context.Context) context.Context {
	return httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GetConn: func(addr string) {
			log.Printf("connection to host and port = %s\n", addr)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			log.Printf("connection acquired: %+v\n", info)
		},
		PutIdleConn: func(err error) {
			if err != nil {
				log.Printf("put idle conn: %+v\n", err)
			}
		},
		GotFirstResponseByte: func() {
			log.Println("got first response byte")
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			log.Printf("dns started for host: %s\n", info.Host)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			log.Printf("dns resvolver done: %+v\n", info)
		},
		ConnectStart: func(network, addr string) {
			log.Printf("connection started at network: %s and addr: %s\n", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			log.Printf("connection done at network: %s and addr: %s\n", network, addr)
		},
		TLSHandshakeStart: func() {
			log.Println("tls handshake started")
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				log.Printf("tls handshake done: %+v\n", state)
			} else {
				log.Println("tls handshake done with err:", err)
			}
		},
		WroteHeaderField: func(key string, value []string) {
			log.Printf("header field written key: %s, value: %v\n", key, value)
		},
		WroteHeaders: func() {
			log.Println("writing of headers completed")
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			log.Printf("writing of request completed: %+v\n", info)
		},
	})
}
