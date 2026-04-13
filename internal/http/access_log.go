package relayhttp

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder 包装 ResponseWriter 以捕获最终 HTTP 状态码，供访问日志使用。
type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

// AccessLog 为每个请求打一条 INFO 日志（方法、路径、状态、耗时），便于本地与生产排障。
// 说明：日志输出到 stdout（与现有 slog JSON 一致），不会在「审计日志」表里重复记录。
func AccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("http_access",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", rec.code,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
