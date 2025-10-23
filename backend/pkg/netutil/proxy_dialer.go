package netutil

import (
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

// dialThroughProxy 通过HTTP代理建立连接
func (d *ProxyDialer) dialThroughProxy(ctx context.Context, network, address string, proxyURL *url.URL) (net.Conn, error) {
	// 连接到代理服务器
	dialer := &net.Dialer{
		Timeout:   d.timeout,
		KeepAlive: d.keepAlive,
	}

	proxyConn, err := dialer.DialContext(ctx, network, proxyURL.Host)
	if err != nil {
		return nil, fmt.Errorf("连接代理服务器失败: %w", err)
	}

	// 发送CONNECT请求
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", address, address)

	// 如果代理需要认证
	if proxyURL.User != nil {
		username := proxyURL.User.Username()
		password, _ := proxyURL.User.Password()
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", auth)
	}

	connectReq += "\r\n"

	// 发送CONNECT请求
	if _, err := proxyConn.Write([]byte(connectReq)); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("发送CONNECT请求失败: %w", err)
	}

	// 读取代理响应
	buffer := make([]byte, 1024)
	n, err := proxyConn.Read(buffer)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("读取代理响应失败: %w", err)
	}

	response := string(buffer[:n])
	if !strings.Contains(response, "200") {
		proxyConn.Close()
		return nil, fmt.Errorf("代理连接失败，响应: %s", response)
	}

	return proxyConn, nil
}

// GetProxyFromEnv 从环境变量获取代理配置信息（用于调试）
func GetProxyFromEnv() (httpProxy, httpsProxy, noProxy string) {
	return os.Getenv("http_proxy"), os.Getenv("https_proxy"), os.Getenv("no_proxy")
}
