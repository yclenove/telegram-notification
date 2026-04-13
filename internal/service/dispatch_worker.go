package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"telegram-notification/internal/config"
	"telegram-notification/internal/model"
	"telegram-notification/internal/repository/postgres"
	relaylegacy "telegram-notification/internal/relay"
	"telegram-notification/internal/telegram"
)

// DispatchWorker 周期扫描待发送任务并投递到 Telegram。
type DispatchWorker struct {
	logger      *slog.Logger
	store       *postgres.Store
	retryCfg    config.RetryConfig
	workerCfg   config.WorkerConfig
	defaultMode string
}

func NewDispatchWorker(logger *slog.Logger, store *postgres.Store, retryCfg config.RetryConfig, workerCfg config.WorkerConfig, defaultMode string) *DispatchWorker {
	return &DispatchWorker{
		logger:      logger,
		store:       store,
		retryCfg:    retryCfg,
		workerCfg:   workerCfg,
		defaultMode: defaultMode,
	}
}

func (w *DispatchWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.workerCfg.PollIntervalMS) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("dispatch worker stopped")
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *DispatchWorker) runOnce(ctx context.Context) {
	jobs, err := w.store.NextPendingJobs(ctx, w.workerCfg.BatchSize)
	if err != nil {
		w.logger.Error("load pending jobs failed", "error", err)
		return
	}
	for _, job := range jobs {
		w.handleJob(ctx, job.ID)
	}
}

func (w *DispatchWorker) handleJob(ctx context.Context, jobID int64) {
	job, event, destination, bot, err := w.store.LoadDispatchContext(ctx, jobID)
	if err != nil {
		w.logger.Error("load dispatch context failed", "job_id", jobID, "error", err)
		return
	}
	client := telegram.NewClient(config.TelegramConfig{
		BotToken:   postgres.DecryptSecret(bot.BotTokenEnc),
		ChatID:     destination.ChatID,
		ParseMode:  destination.ParseMode,
		APIBaseURL: "https://api.telegram.org",
		TimeoutSec: 5,
	})
	service := relaylegacy.NewService(client, w.retryCfg)
	err = service.Send(ctx, model.NotifyRequest{
		Title:   event.Title,
		Message: event.Message,
		Level:   event.Level,
		Source:  event.Source,
		EventID: event.EventID,
	})
	if err == nil {
		if err = w.store.MarkJobSuccess(ctx, job.ID, event.ID); err != nil {
			w.logger.Error("mark job success failed", "job_id", job.ID, "error", err)
		}
		return
	}
	backoff := time.Duration(w.retryCfg.InitialBackoffMS) * time.Millisecond
	if err2 := w.store.MarkJobFailedOrRetry(ctx, job, event.ID, err.Error(), backoff); err2 != nil {
		w.logger.Error("mark job failed/retry failed", "job_id", job.ID, "error", err2)
	}
	w.logger.Error("dispatch job failed", "job_id", job.ID, "event_id", event.EventID, "error", fmt.Sprintf("%v", err))
}
