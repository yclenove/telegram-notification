package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yclenove/telegram-relay/internal/config"
)

type Client struct {
	httpClient *http.Client
	botToken   string
	chatID     string
	parseMode  string
	baseURL    string
}

type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// NewClient 创建 Telegram API 客户端；若配置了 telegram.proxy_url / TELEGRAM_PROXY，则经该代理访问 Bot API。
func NewClient(cfg config.TelegramConfig) (*Client, error) {
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	transport, err := newHTTPTransportForProxy(cfg.ProxyURL)
	if err != nil {
		return nil, err
	}
	var httpClient *http.Client
	if transport != nil {
		httpClient = &http.Client{Transport: transport, Timeout: timeout}
	} else {
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{
		httpClient: httpClient,
		botToken:   cfg.BotToken,
		chatID:     cfg.ChatID,
		parseMode:  cfg.ParseMode,
		baseURL:    strings.TrimRight(cfg.APIBaseURL, "/"),
	}, nil
}

// Send 将文本消息发送到指定 chat_id。
// 调用失败会返回明确的上下文错误，便于上层重试和记录日志。
func (c *Client) Send(ctx context.Context, text string) error {
	body, err := json.Marshal(sendMessageRequest{
		ChatID:    c.chatID,
		Text:      text,
		ParseMode: c.parseMode,
	})
	if err != nil {
		return fmt.Errorf("marshal telegram request: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("telegram api error: status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}
