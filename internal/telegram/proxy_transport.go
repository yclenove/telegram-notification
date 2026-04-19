package telegram

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"
)

// newHTTPTransportForProxy 根据可选代理 URL 构造 *http.Transport；proxyURL 为空时返回 (nil, nil) 表示使用 http.Client 默认 Transport。
// 支持 http/https（如 v2rayN 本地 HTTP 入站，常见端口为 SOCKS+1，即 10809）与 socks5/socks5h（如 10808）。
func newHTTPTransportForProxy(proxyURL string) (*http.Transport, error) {
	raw := strings.TrimSpace(proxyURL)
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("telegram proxy url: %w", err)
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		tr.Proxy = http.ProxyURL(u)
		return tr, nil
	case "socks5", "socks5h":
		dialer, err := proxy.FromURL(u, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("telegram socks proxy: %w", err)
		}
		if cd, ok := dialer.(proxy.ContextDialer); ok {
			tr.DialContext = cd.DialContext
		} else {
			tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				_ = ctx
				return dialer.Dial(network, addr)
			}
		}
		tr.Proxy = nil
		return tr, nil
	default:
		return nil, fmt.Errorf("telegram proxy: unsupported scheme %q", u.Scheme)
	}
}
