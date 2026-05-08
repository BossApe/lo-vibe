package external

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"musuhi-api/internal/handler"
	"musuhi-api/internal/middleware"
	"musuhi-api/internal/repository"
	"musuhi-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type overviewIDEnvelope struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type extractionEnvelope struct {
	Data struct {
		Features   []string `json:"features"`
		Components []string `json:"components"`
	} `json:"data"`
}

type suggestionEnvelope struct {
	Data struct {
		Candidates []string `json:"candidates"`
	} `json:"data"`
}

func newExternalProjectServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()
	pool := newExternalTestPool(t)
	soRepo := repository.NewPostgresSystemOverviewRepository(pool)
	soSvc := service.NewSystemOverviewService(soRepo)
	projectSvc := service.NewProjectService(soRepo)
	soHandler := handler.NewSystemOverviewHandler(soSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/system-overviews", soHandler.Create)
	mux.HandleFunc("POST /api/v1/projects/extract-features", projectHandler.ExtractFeatures)
	mux.HandleFunc("POST /api/v1/projects/suggest-name", projectHandler.SuggestName)
	mux.HandleFunc("POST /api/v1/projects/init-directory", projectHandler.InitDirectory)
	srv := httptest.NewServer(middleware.Logger(middleware.CORS(mux)))
	return srv, func() { srv.Close() }
}

// createTestOverview は外部テスト用にシステム概要を実DBに登録してIDを返す。
func createTestOverview(t *testing.T, srvURL string, content string) string {
	t.Helper()
	body := `{"content":"` + content + `"}`
	res, err := http.Post(srvURL+"/api/v1/system-overviews", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusCreated, res.StatusCode)
	var resp overviewIDEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	return resp.Data.ID
}

func TestExternalIntegration_ExtractFeatures_実DBのシステム概要から機能一覧を抽出する_正常系(t *testing.T) {
	srv, cleanup := newExternalProjectServer(t)
	defer cleanup()

	overviewID := createTestOverview(t, srv.URL, "ユーザー管理機能とレポート出力機能を持つWebシステム")

	body, err := json.Marshal(map[string]string{"overviewId": overviewID})
	require.NoError(t, err)
	res, err := http.Post(srv.URL+"/api/v1/projects/extract-features", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	var resp extractionEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	assert.NotNil(t, resp.Data.Features)
	assert.NotNil(t, resp.Data.Components)
}

func TestExternalIntegration_SuggestName_実DBのシステム概要からプロジェクト名候補を取得する_正常系(t *testing.T) {
	srv, cleanup := newExternalProjectServer(t)
	defer cleanup()

	overviewID := createTestOverview(t, srv.URL, "在庫管理と発注管理を行うECシステム")

	body, err := json.Marshal(map[string]string{"overviewId": overviewID})
	require.NoError(t, err)
	res, err := http.Post(srv.URL+"/api/v1/projects/suggest-name", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	var resp suggestionEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Data.Candidates)
}

func TestExternalIntegration_ExtractFeatures_存在しない概要IDを指定する_異常系(t *testing.T) {
	srv, cleanup := newExternalProjectServer(t)
	defer cleanup()

	body, err := json.Marshal(map[string]string{"overviewId": "00000000-0000-0000-0000-000000000000"})
	require.NoError(t, err)
	res, err := http.Post(srv.URL+"/api/v1/projects/extract-features", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestExternalIntegration_InitDirectory_有効なパスに初期ディレクトリを作成する_正常系(t *testing.T) {
	srv, cleanup := newExternalProjectServer(t)
	defer cleanup()

	tmpDir := t.TempDir()
	body, err := json.Marshal(map[string]string{
		"projectName": "test-project",
		"localPath":   tmpDir,
		"template":    "default",
	})
	require.NoError(t, err)
	res, err := http.Post(srv.URL+"/api/v1/projects/init-directory", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	dirInfo, statErr := os.Stat(tmpDir + "/_document")
	require.NoError(t, statErr)
	assert.True(t, dirInfo.IsDir())
	_, readmeErr := os.Stat(tmpDir + "/README.md")
	assert.NoError(t, readmeErr)
}
