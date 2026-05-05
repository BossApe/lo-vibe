package src

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"musuhi-api/internal/handler"
	"musuhi-api/internal/middleware"
	"musuhi-api/internal/repository"
	"musuhi-api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 構造化ロガー設定
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// DB接続
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://musuhi:musuhi@localhost:5432/musuhi?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 接続確認（最大10秒リトライ）
	for i := range 10 {
		if pingErr := pool.Ping(ctx); pingErr == nil {
			break
		}
		slog.Info("waiting for database...", "attempt", i+1)
		time.Sleep(time.Second)
	}

	// 依存性注入
	soRepo := repository.NewPostgresSystemOverviewRepository(pool)
	soSvc := service.NewSystemOverviewService(soRepo)
	soHandler := handler.NewSystemOverviewHandler(soSvc)

	// ルーティング（Go 1.22 enhanced ServeMux）
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /api/v1/system-overviews", soHandler.Create)
	mux.HandleFunc("GET /api/v1/system-overviews/{id}", soHandler.GetByID)

	// ミドルウェアチェーン
	h := middleware.Logger(middleware.CORS(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("musuhi-api starting", "port", port)
	if err := http.ListenAndServe(":"+port, h); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
