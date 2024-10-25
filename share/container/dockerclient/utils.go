package dockerclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

type tcpFunc func(*net.TCPConn, time.Duration) error

func newHTTPClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration, setUserTimeout tcpFunc) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	default:
		httpTransport.DialContext = func(ctx context.Context, proto, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(proto, addr, timeout)
			if tcpConn, ok := conn.(*net.TCPConn); ok && setUserTimeout != nil {
				// Sender can break TCP connection if the remote side doesn't
				// acknowledge packets within timeout
				_ = setUserTimeout(tcpConn, timeout)
			}
			return conn, err
		}
	case "unix":
		socketPath := u.Path
		httpTransport.DialContext = func(ctx context.Context, proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}
}
