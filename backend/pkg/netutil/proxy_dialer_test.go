package netutil

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestProxyDialer_DirectConnection(t *testing.T) {
	const payload = "direct-connection-ok"

	target, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听目标端口失败: %v", err)
	}
	defer target.Close()

	go func() {
		conn, err := target.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte(payload))
	}()

	dialer := &ProxyDialer{
		timeout:   time.Second,
		keepAlive: time.Second,
		proxy: func(*http.Request) (*url.URL, error) {
			return nil, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", target.Addr().String())
	if err != nil {
		t.Fatalf("直连拨号失败: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))

	buf := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("读取直连返回数据失败: %v", err)
	}

	if string(buf) != payload {
		t.Fatalf("直连返回数据不一致: got=%q want=%q", string(buf), payload)
	}
}

func TestProxyDialer_HTTPProxy(t *testing.T) {
	const proxyResponse = "http-proxy-response"

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听 HTTP 代理失败: %v", err)
	}
	defer proxyListener.Close()

	targetAddr := "imap.example.com:993"

	go func() {
		conn, err := proxyListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			t.Errorf("代理读取 CONNECT 请求失败: %v", err)
			return
		}
		if req.Method != http.MethodConnect {
			t.Errorf("CONNECT 方法错误: got=%s", req.Method)
		}
		if req.Host != targetAddr {
			t.Errorf("CONNECT 目标错误: got=%s want=%s", req.Host, targetAddr)
		}
		_, _ = fmt.Fprintf(conn, "HTTP/1.1 200 Connection established\r\n\r\n")
		_, _ = conn.Write([]byte(proxyResponse))
	}()

	dialer := &ProxyDialer{
		timeout:   time.Second,
		keepAlive: time.Second,
		proxy: func(*http.Request) (*url.URL, error) {
			return &url.URL{
				Scheme: "http",
				Host:   proxyListener.Addr().String(),
			}, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		t.Fatalf("HTTP 代理拨号失败: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))

	buf := make([]byte, len(proxyResponse))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("读取 HTTP 代理返回数据失败: %v", err)
	}
	if string(buf) != proxyResponse {
		t.Fatalf("HTTP 代理返回数据不一致: got=%q want=%q", string(buf), proxyResponse)
	}
}

func TestProxyDialer_HTTPSProxy(t *testing.T) {
	const proxyResponse = "https-proxy-response"

	cert := mustGenerateSelfSignedCert(t, "127.0.0.1")

	tlsListener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		t.Fatalf("监听 HTTPS 代理失败: %v", err)
	}
	defer tlsListener.Close()

	targetAddr := "imap.example.com:993"

	go func() {
		conn, err := tlsListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			t.Errorf("HTTPS 代理读取 CONNECT 请求失败: %v", err)
			return
		}
		if req.Method != http.MethodConnect {
			t.Errorf("CONNECT 方法错误: got=%s", req.Method)
		}
		if req.Host != targetAddr {
			t.Errorf("CONNECT 目标错误: got=%s want=%s", req.Host, targetAddr)
		}
		_, _ = fmt.Fprintf(conn, "HTTP/1.1 200 Connection established\r\n\r\n")
		_, _ = conn.Write([]byte(proxyResponse))
	}()

	restore := setProxyTLSConfigFactory(func(serverName string) *tls.Config {
		return &tls.Config{
			ServerName:         serverName,
			InsecureSkipVerify: true,
		}
	})
	defer restore()

	dialer := &ProxyDialer{
		timeout:   time.Second,
		keepAlive: time.Second,
		proxy: func(*http.Request) (*url.URL, error) {
			return &url.URL{
				Scheme: "https",
				Host:   tlsListener.Addr().String(),
			}, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		t.Fatalf("HTTPS 代理拨号失败: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))

	buf := make([]byte, len(proxyResponse))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("读取 HTTPS 代理返回数据失败: %v", err)
	}
	if string(buf) != proxyResponse {
		t.Fatalf("HTTPS 代理返回数据不一致: got=%q want=%q", string(buf), proxyResponse)
	}
}

func mustGenerateSelfSignedCert(t *testing.T, host string) tls.Certificate {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成证书私钥失败: %v", err)
	}

	notBefore := time.Now().Add(-time.Minute)
	notAfter := notBefore.Add(24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = []net.IP{ip}
	} else {
		template.DNSNames = []string{host}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("生成自签证书失败: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("解析证书失败: %v", err)
	}

	return cert
}

func setProxyTLSConfigFactory(factory func(string) *tls.Config) func() {
	prev := newProxyTLSConfig
	newProxyTLSConfig = factory
	return func() {
		newProxyTLSConfig = prev
	}
}
