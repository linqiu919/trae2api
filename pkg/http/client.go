package http

import (
	"crypto/tls"
	"io"
	"net/http"
	"time"

	"github.com/trae2api/pkg/logger"
)

// DebugTransport 是一个调试用的Transport包装器
type DebugTransport struct {
	Transport http.RoundTripper
}

// RoundTrip 实现http.RoundTripper接口
func (d *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 记录请求的协议版本
	logger.Log.Debugf("请求协议: %s, URL: %s", req.Proto, req.URL.String())

	// 执行实际请求
	resp, err := d.Transport.RoundTrip(req)

	// 记录响应的协议版本
	if err == nil && resp != nil {
		logger.Log.Debugf("响应协议: %s, 状态码: %d", resp.Proto, resp.StatusCode)
	}

	return resp, err
}

// NewHTTP11Client 创建一个严格使用HTTP/1.1的HTTP客户端
func NewHTTP11Client() *http.Client {
	// 创建一个配置为HTTP/1.1的传输
	transport := &http.Transport{
		DisableKeepAlives: false,
		ForceAttemptHTTP2: false,                                                                  // 强制不使用HTTP/2
		TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) http.RoundTripper), // 禁用HTTP/2
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS13,
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false,
	}

	// 包装调试Transport
	debugTransport := &DebugTransport{
		Transport: transport,
	}

	return &http.Client{
		Transport: debugTransport,
		Timeout:   30 * time.Second,
	}
}

// NewHTTP11Request 创建一个使用HTTP/1.1的请求
func NewHTTP11Request(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// 确保使用HTTP/1.1
	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1

	return req, nil
}
