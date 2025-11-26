package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Shared HTTP transport for outbound requests
var transport = &http.Transport{
	Proxy:               nil, // no upstream proxy; this *is* the proxy
	MaxIdleConns:        100,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 10 * time.Second,
}

func main() {
	r := gin.Default()

	// Single handler for all paths
	r.Any("/*proxyPath", proxyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Starting forward proxy on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// proxyHandler forwards incoming requests and returns the upstream response.
func proxyHandler(c *gin.Context) {
	req := c.Request

	if strings.ToUpper(req.Method) == http.MethodConnect {
		handleConnect(c)
		return
	}

	handleHTTP(c)
}

// handleHTTP handles normal HTTP methods (GET/POST/etc.) in forward-proxy style.
func handleHTTP(c *gin.Context) {
	inReq := c.Request

	// FIX: Remove the leading slash from the proxy path
	rawTarget := strings.TrimPrefix(inReq.RequestURI, "/")
	targetURL, err := url.Parse(rawTarget)
	if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
		log.Printf("invalid absolute target URL: %v", err)
		c.String(http.StatusBadRequest, "Bad target URL")
		return
	}

	outReq, err := http.NewRequest(inReq.Method, targetURL.String(), inReq.Body)
	if err != nil {
		log.Printf("failed to create outbound request: %v", err)
		c.String(http.StatusInternalServerError, "Failed to create outbound request")
		return
	}

	// Copy and sanitize headers
	copyHeaders(outReq.Header, inReq.Header)
	removeHopByHop(outReq.Header)

	outReq = outReq.WithContext(c.Request.Context())

	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		log.Printf("proxy upstream error: %v", err)
		c.String(http.StatusBadGateway, "Upstream request failed")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close response body: %v", err)
			c.String(http.StatusInternalServerError, "Failed to close response body")
			return
		}
	}(resp.Body)

	// Write headers + body back to client
	for k, vv := range resp.Header {
		for _, v := range vv {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

// handleConnect supports HTTPS tunneling (CONNECT method).
func handleConnect(c *gin.Context) {
	hijacker, ok := c.Writer.(http.Hijacker)
	if !ok {
		http.Error(c.Writer, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, buf, err := hijacker.Hijack()
	if err != nil {
		log.Printf("hijack error: %v", err)
		return
	}
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Printf("failed to close client connection: %v", err)
			return
		}
	}(clientConn)

	host := c.Request.Host // e.g. "example.com:443"

	// Dial the target host
	targetConn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		log.Printf("dial to %s failed: %v", host, err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	// targetConn closed via defer in copy
	defer func(targetConn net.Conn) {
		err := targetConn.Close()
		if err != nil {
			log.Printf("failed to close target connection: %v", err)
			return
		}
	}(targetConn)

	// Tell client the tunnel is established
	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// If the client already sent buffered data, forward it
	go func() {
		if buf.Reader.Buffered() > 0 {
			_, _ = io.Copy(targetConn, buf)
		}
	}()

	// Bidirectional copy
	errCh := make(chan struct{}, 2)

	go func() {
		_, _ = io.Copy(targetConn, clientConn)
		errCh <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(clientConn, targetConn)
		errCh <- struct{}{}
	}()

	// Wait for one side to close
	<-errCh
}

// copyHeaders clones headers from src to dst.
func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// removeHopByHop removes hop-by-hop headers not to be forwarded.
func removeHopByHop(h http.Header) {
	h.Del("Connection")
	h.Del("Proxy-Connection")
	h.Del("Keep-Alive")
	h.Del("Proxy-Authenticate")
	h.Del("Proxy-Authorization")
	h.Del("TE")
	h.Del("Trailers")
	h.Del("Transfer-Encoding")
	h.Del("Upgrade")
}
