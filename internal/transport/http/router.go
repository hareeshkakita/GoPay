package httptransport

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouterConfig struct {
	Logger  *slog.Logger
	Handler *Handler
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	if cfg.Logger != nil {
		r.Use(loggingMiddleware(cfg.Logger))
	}

	r.Get("/health", healthHandler)
	r.Get("/ready", readinessHandler)
	r.Post("/wallets", cfg.Handler.CreateWallet)
	r.Get("/wallets/{walletID}", cfg.Handler.GetWallet)
	r.Get("/wallets/{walletID}/balance", cfg.Handler.GetBalance)
	r.Post("/wallets/{walletID}/deposit", cfg.Handler.DepositMoney)
	r.Post("/wallets/{walletID}/withdraw", cfg.Handler.WithdrawMoney)
	r.Post("/wallets/transfer", cfg.Handler.TransferMoney)

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ready")
}

func loggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info(
				"request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}
