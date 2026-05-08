//go:build external

// Package external は実際の PostgreSQL DB に接続する外部結合テストです。
// 実行するには Docker Compose でサービスを起動し、DATABASE_URL を設定した上で
// go test -tags external ./test/external/... -v
// を実行してください。
package external

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"musuhi-api/internal/handler"
	"musuhi-api/internal/middleware"
	"musuhi-api/internal/repository"
	"musuhi-api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type systemOverviewEnvelope struct {
	Data struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	} `json:"data"`
}

func newExternalTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://musuhi:musuhi@localhost:5432/musuhi?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "DB接続失敗: DATABASE_URL=%s", dsn)
	require.NoError(t, pool.Ping(ctx), "DB疎通確認失敗")
	t.Cleanup(pool.Close)
	return pool
}

func newExternalSystemOverviewServer(pool *pgxpool.Pool) *httptest.Server {
	soRepo := repository.NewPostgresSystemOverviewRepository(pool)
	soSvc := service.NewSystemOverviewService(soRepo)
	soHandler := handler.NewSystemOverviewHandler(soSvc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/system-overviews", soHandler.Create)
	mux.HandleFunc("GET /api/v1/system-overviews/{id}", soHandler.GetByID)
	mux.HandleFunc("PUT /api/v1/system-overviews/{id}", soHandler.Update)
	return httptest.NewServer(middleware.Logger(middleware.CORS(mux)))
}

func TestExternalIntegration_CreateSystemOverview_実DBにシステム概要を保存する_正常系(t *testing.T) {
	pool := newExternalTestPool(t)
	srv := newExternalSystemOverviewServer(pool)
	defer srv.Close()

	body := `{"content":"外部結合テスト用システム概要"}`
	res, err := http.Post(srv.URL+"/api/v1/system-overviews", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))

	var resp systemOverviewEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Data.ID)
	assert.Equal(t, "外部結合テスト用システム概要", resp.Data.Content)
}

func TestExternalIntegration_GetSystemOverviewByID_実DBからシステム概要を取得する_正常系(t *testing.T) {
	pool := newExternalTestPool(t)
	srv := newExternalSystemOverviewServer(pool)
	defer srv.Close()

	// 事前にデータ登録
	body := `{"content":"取得テスト用概要"}`
	createRes, err := http.Post(srv.URL+"/api/v1/system-overviews", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer createRes.Body.Close()
	require.Equal(t, http.StatusCreated, createRes.StatusCode)

	var created systemOverviewEnvelope
	require.NoError(t, json.NewDecoder(createRes.Body).Decode(&created))
	id := created.Data.ID

	// 取得
	res, err := http.Get(srv.URL + "/api/v1/system-overviews/" + id)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	var resp systemOverviewEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	assert.Equal(t, id, resp.Data.ID)
	assert.Equal(t, "取得テスト用概要", resp.Data.Content)
}

func TestExternalIntegration_UpdateSystemOverview_実DBのシステム概要を更新する_正常系(t *testing.T) {
	pool := newExternalTestPool(t)
	srv := newExternalSystemOverviewServer(pool)
	defer srv.Close()

	// 事前にデータ登録
	body := `{"content":"更新前概要"}`
	createRes, err := http.Post(srv.URL+"/api/v1/system-overviews", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer createRes.Body.Close()
	require.Equal(t, http.StatusCreated, createRes.StatusCode)

	var created systemOverviewEnvelope
	require.NoError(t, json.NewDecoder(createRes.Body).Decode(&created))
	id := created.Data.ID

	// 更新
	updateBody, err := json.Marshal(map[string]string{"content": "更新後概要"})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/system-overviews/"+id, bytes.NewReader(updateBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	var resp systemOverviewEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	assert.Equal(t, id, resp.Data.ID)
	assert.Equal(t, "更新後概要", resp.Data.Content)
}

func TestExternalIntegration_GetSystemOverviewByID_存在しないUUIDで取得する_異常系(t *testing.T) {
	pool := newExternalTestPool(t)
	srv := newExternalSystemOverviewServer(pool)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/v1/system-overviews/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
