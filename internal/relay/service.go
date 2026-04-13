package relay

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yclenove/telegram-relay/internal/config"
	"github.com/yclenove/telegram-relay/internal/model"
	"github.com/yclenove/telegram-relay/internal/retry"
	"github.com/yclenove/telegram-relay/internal/telegram"
)

type Service struct {
	client   *telegram.Client
	retryCfg retry.Config
}

// NewService 构建中转服务并将配置转换为 retry 层可用结构。
func NewService(client *telegram.Client, cfg config.RetryConfig) *Service {
	return &Service{
		client: client,
		retryCfg: retry.Config{
			MaxAttempts:    cfg.MaxAttempts,
			InitialBackoff: time.Duration(cfg.InitialBackoffMS) * time.Millisecond,
			MaxBackoff:     time.Duration(cfg.MaxBackoffMS) * time.Millisecond,
		},
	}
}

// Send 负责完成“格式化消息 + 重试发送”流程。
func (s *Service) Send(ctx context.Context, req model.NotifyRequest) error {
	msg := formatMessage(req)
	return retry.Do(ctx, s.retryCfg, func() error {
		return s.client.Send(ctx, msg)
	})
}

// formatMessage 将统一告警模型转换为 Telegram HTML 消息文本。
func formatMessage(req model.NotifyRequest) string {
	builder := strings.Builder{}
	builder.WriteString("<b>告警通知</b>\n")
	builder.WriteString(fmt.Sprintf("<b>标题:</b> %s\n", escapeHTML(req.Title)))
	builder.WriteString(fmt.Sprintf("<b>级别:</b> %s\n", escapeHTML(req.Level)))
	builder.WriteString(fmt.Sprintf("<b>来源:</b> %s\n", escapeHTML(req.Source)))
	builder.WriteString(fmt.Sprintf("<b>事件ID:</b> %s\n", escapeHTML(req.EventID)))
	if req.EventTime != "" {
		builder.WriteString(fmt.Sprintf("<b>时间:</b> %s\n", escapeHTML(req.EventTime)))
	}
	builder.WriteString(fmt.Sprintf("<b>内容:</b>\n%s\n", escapeHTML(req.Message)))

	if len(req.Labels) > 0 {
		builder.WriteString("<b>标签:</b>\n")
		keys := make([]string, 0, len(req.Labels))
		for k := range req.Labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			builder.WriteString(fmt.Sprintf("- %s=%s\n", escapeHTML(k), escapeHTML(req.Labels[k])))
		}
	}
	return builder.String()
}

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#39;",
)

func escapeHTML(s string) string {
	return htmlReplacer.Replace(s)
}
