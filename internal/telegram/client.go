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

// NewClient 创建 Telegram API 客户端。
func NewClient(cfg config.TelegramConfig) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: time.Duration(cfg.TimeoutSec) * time.Second},
		botToken:   cfg.BotToken,
		chatID:     cfg.ChatID,
		parseMode:  cfg.ParseMode,
		baseURL:    strings.TrimRight(cfg.APIBaseURL, "/"),
	}
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
