package netutil

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ProxyDialer 支持HTTP代理的拨号器
type ProxyDialer struct {
	timeout   time.Duration
	keepAlive time.Duration
	proxy     func(*http.Request) (*url.URL, error)
}

var newProxyTLSConfig = func(serverName string) *tls.Config {
	return &tls.Config{
		ServerName: serverName,
	}
}

// NewProxyDialer 创建新的代理拨号器
func NewProxyDialer(timeout, keepAlive time.Duration) *ProxyDialer {
	return &ProxyDialer{
		timeout:   timeout,
		keepAlive: keepAlive,
		proxy:     http.ProxyFromEnvironment,
	}
}

// DialContext 使用代理进行网络连接
func (d *ProxyDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// 创建一个虚拟的HTTP请求来获取代理配置
	req := &http.Request{
		Method: "CONNECT",
		URL: &url.URL{
			Scheme: "http",
			Host:   address,
		},
		Header: make(http.Header),
	}

	proxyURL, err := d.proxy(req)
	if err != nil {
		return nil, fmt.Errorf("获取代理配置失败: %w", err)
	}

	// 如果没有配置代理，直接连接
	if proxyURL == nil {
		dialer := &net.Dialer{
			Timeout:   d.timeout,
			KeepAlive: d.keepAlive,
		}
		return dialer.DialContext(ctx, network, address)
	}

	// 使用HTTP代理连接
	return d.dialThroughProxy(ctx, network, address, proxyURL)
}

// DialTLSContext 使用代理进行TLS连接
func (d *ProxyDialer) DialTLSContext(ctx context.Context, network, address string, tlsConfig *tls.Config) (net.Conn, error) {
	// 创建一个虚拟的HTTP请求来获取代理配置
	req := &http.Request{
		Method: "CONNECT",
		URL: &url.URL{
			Scheme: "https",
			Host:   address,
		},
		Header: make(http.Header),
	}

	proxyURL, err := d.proxy(req)
	if err != nil {
		return nil, fmt.Errorf("获取代理配置失败: %w", err)
	}

	// 如果没有配置代理，直接TLS连接
	if proxyURL == nil {
		dialer := &net.Dialer{
			Timeout:   d.timeout,
			KeepAlive: d.keepAlive,
		}
		conn, err := dialer.DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, fmt.Errorf("TLS握手失败: %w", err)
		}
		return tlsConn, nil
	}

	// 通过代理建立连接
	conn, err := d.dialThroughProxy(ctx, network, address, proxyURL)
	if err != nil {
		return nil, err
	}

	// 在代理连接上进行TLS握手
	tlsConn := tls.Client(conn, tlsConfig)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("TLS握手失败: %w", err)
	}

	return tlsConn, nil
}

// dialThroughProxy 通过代理建立 CONNECT 通道
func (d *ProxyDialer) dialThroughProxy(ctx context.Context, network, address string, proxyURL *url.URL) (net.Conn, error) {
	host := proxyURL.Hostname()
	port := proxyURL.Port()
	if port == "" {
		if strings.EqualFold(proxyURL.Scheme, "https") {
			port = "443"
		} else {
			port = "80"
		}
	}
	proxyAddress := net.JoinHostPort(host, port)

	// 连接到代理服务器
	dialer := &net.Dialer{
		Timeout:   d.timeout,
		KeepAlive: d.keepAlive,
	}

	baseConn, err := dialer.DialContext(ctx, network, proxyAddress)
	if err != nil {
		return nil, fmt.Errorf("连接代理服务器失败: %w", err)
	}

	proxyConn := net.Conn(baseConn)

	if strings.EqualFold(proxyURL.Scheme, "https") {
		tlsConfig := newProxyTLSConfig(host)
		if tlsConfig == nil {
			baseConn.Close()
			return nil, fmt.Errorf("HTTPS 代理缺少 TLS 配置")
		}
		tlsConn := tls.Client(baseConn, tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			tlsConn.Close()
			return nil, fmt.Errorf("HTTPS 代理握手失败: %w", err)
		}
		proxyConn = tlsConn
	} else if proxyURL.Scheme != "" && !strings.EqualFold(proxyURL.Scheme, "http") {
		baseConn.Close()
		return nil, fmt.Errorf("不支持的代理协议: %s", proxyURL.Scheme)
	}

	if d.timeout > 0 {
		deadline := time.Now().Add(d.timeout)
		_ = proxyConn.SetDeadline(deadline)
		defer func() {
			_ = proxyConn.SetDeadline(time.Time{})
		}()
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Opaque: address,
		},
		Host:   address,
		Header: make(http.Header),
	}

	// 如果代理需要认证
	if proxyURL.User != nil {
		username := proxyURL.User.Username()
		password, _ := proxyURL.User.Password()
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		req.Header.Set("Proxy-Authorization", fmt.Sprintf("Basic %s", auth))
	}

	// 发送CONNECT请求
	if err := req.Write(proxyConn); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("发送CONNECT请求失败: %w", err)
	}

	// 读取代理响应
	reader := bufio.NewReader(proxyConn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("读取代理响应失败: %w", err)
	}

	// CONNECT 响应不应带有正文，关闭 Body 以复用连接
	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		proxyConn.Close()
		return nil, fmt.Errorf("代理连接失败，状态码: %d", resp.StatusCode)
	}

	return &bufferedProxyConn{
		Conn:   proxyConn,
		reader: reader,
	}, nil
}

// GetProxyFromEnv 从环境变量获取代理配置信息（用于调试）
func GetProxyFromEnv() (httpProxy, httpsProxy, noProxy string) {
	return os.Getenv("http_proxy"), os.Getenv("https_proxy"), os.Getenv("no_proxy")
}

// bufferedProxyConn 在 CONNECT 握手后保留已读取但未消费的数据
type bufferedProxyConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedProxyConn) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}
