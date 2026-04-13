package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"telegram-notification/internal/model"
	"telegram-notification/internal/repository/postgres"
)

// NotifyService 负责把入站告警转换为事件+任务。
type NotifyService struct {
	store       *postgres.Store
	maxAttempts int
}

func NewNotifyService(store *postgres.Store, maxAttempts int) *NotifyService {
	return &NotifyService{store: store, maxAttempts: maxAttempts}
}

func (s *NotifyService) Ingest(ctx context.Context, req model.NotifyRequest, rawBody []byte) (int64, error) {
	// 当上游没给 event_id 时，本地兜底生成幂等键。
	if req.EventID == "" {
		req.EventID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	destination, err := s.store.ResolveDestinationByRules(ctx, req)
	if err != nil {
		return 0, err
	}
	return s.store.CreateEventAndJob(ctx, req, string(rawBody), destination.ID, s.maxAttempts)
}

func BuildAuditDetail(data any) string {
	raw, _ := json.Marshal(data)
	return string(raw)
}
