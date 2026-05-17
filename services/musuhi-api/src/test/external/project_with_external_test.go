package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"musuhi-api/internal/handler"
	"musuhi-api/internal/middleware"
	"musuhi-api/internal/repository"
	"musuhi-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type withExternalResponseEnvelope struct {
	Data struct {
		RepositoryURL     string `json:"repositoryUrl"`
		ExternalProjectID string `json:"externalProjectId"`
		PushStatus        string `json:"pushStatus"`
	} `json:"data"`
}

func requireGitHubExternalTestOwner(t *testing.T) string {
	t.Helper()
	owner := strings.TrimSpace(os.Getenv("GITHUB_TEST_OWNER"))
	if owner == "" {
		owner = strings.TrimSpace(os.Getenv("GH_TEST_OWNER"))
	}
	if owner == "" {
		t.Skip("GITHUB_TEST_OWNER もしくは GH_TEST_OWNER を設定してください")
	}
	return owner
}

func ensureGHAuth(t *testing.T) {
	t.Helper()
	cmd := exec.Command("gh", "auth", "status")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "gh auth status failed: %s", strings.TrimSpace(string(out)))
}

func hasDeleteRepoScope(t *testing.T) bool {
	t.Helper()
	cmd := exec.Command("gh", "auth", "status")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "gh auth status failed: %s", strings.TrimSpace(string(out)))
	return strings.Contains(strings.ToLower(string(out)), "delete_repo")
}

func deleteGitHubRepoIfExists(t *testing.T, fullName string) {
	t.Helper()
	cmd := exec.Command("gh", "repo", "delete", fullName, "--yes")
	if out, err := cmd.CombinedOutput(); err != nil {
		msg := strings.ToLower(string(out))
		if !strings.Contains(msg, "not found") &&
			!strings.Contains(msg, "delete_repo") &&
			!strings.Contains(msg, "must have admin rights") {
			t.Logf("cleanup warning: gh repo delete failed for %s: %s", fullName, strings.TrimSpace(string(out)))
		}
	}
}

func assertGitHubRepoDeleted(t *testing.T, fullName string) {
	t.Helper()
	cmd := exec.Command("gh", "repo", "view", fullName, "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("repository still exists after cleanup: %s", strings.TrimSpace(string(out)))
	}
	msg := strings.ToLower(string(out))
	assert.True(
		t,
		strings.Contains(msg, "not found") || strings.Contains(msg, "could not resolve to a repository"),
		"unexpected gh repo view error: %s",
		strings.TrimSpace(string(out)),
	)
}

func TestExternalIntegration_WithExternal_実GitHubにリポジトリ作成して初回pushする_正常系(t *testing.T) {
	if os.Getenv("RUN_EXTERNAL_TESTS") != "1" {
		t.Skip("external integration tests are disabled; set RUN_EXTERNAL_TESTS=1 to run")
	}

	owner := requireGitHubExternalTestOwner(t)
	ensureGHAuth(t)
	canDeleteRepo := hasDeleteRepoScope(t)

	pool := newExternalTestPool(t)
	soRepo := repository.NewPostgresSystemOverviewRepository(pool)
	projectSvc := service.NewProjectService(soRepo)
	projectHandler := handler.NewProjectHandler(projectSvc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/projects/with-external", projectHandler.WithExternal)
	srv := httptest.NewServer(middleware.Logger(middleware.CORS(mux)))
	defer srv.Close()

	repoName := fmt.Sprintf("musuhi-fr003-e2e-%d", time.Now().Unix())
	fullName := owner + "/" + repoName
	if canDeleteRepo {
		deleteGitHubRepoIfExists(t, fullName)
		t.Cleanup(func() {
			deleteGitHubRepoIfExists(t, fullName)
			assertGitHubRepoDeleted(t, fullName)
			t.Logf("cleanup completed: deleted repository %s", fullName)
		})
	} else {
		t.Log("delete_repo scope is missing; repository cleanup verification is skipped")
	}

	localPath := filepath.Join(t.TempDir(), repoName)
	require.NoError(t, os.MkdirAll(localPath, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(localPath, "README.md"), []byte("# FR-003 external integration\n"), 0o644))

	reqBody, err := json.Marshal(map[string]string{
		"owner":         owner,
		"repoName":      repoName,
		"visibility":    "private",
		"localPath":     localPath,
		"commitMessage": "test: external integration initial commit",
	})
	require.NoError(t, err)

	res, err := http.Post(srv.URL+"/api/v1/projects/with-external", "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp withExternalResponseEnvelope
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	assert.Equal(t, "success", resp.Data.PushStatus)
	assert.NotEmpty(t, resp.Data.RepositoryURL)
	assert.Contains(t, resp.Data.RepositoryURL, fullName)

	viewCmd := exec.Command("gh", "repo", "view", fullName, "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	viewOut, viewErr := viewCmd.CombinedOutput()
	require.NoError(t, viewErr, "gh repo view failed: %s", strings.TrimSpace(string(viewOut)))
	assert.Equal(t, fullName, strings.TrimSpace(string(viewOut)))
}
