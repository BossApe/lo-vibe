package service

// ExternalGitClient は Forgejo（Gitea互換）リポジトリ作成と initial push の実行を抽象化するインターフェース。

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"musuhi-api/internal/model"
)

// GitHubClient は GitHub リポジトリ作成と initial push の実行を抽象化する。
type ExternalGitClient interface {
	CreateRepositoryAndInitialPush(ctx context.Context, owner, repoName, visibility, localPath, commitMessage string) (*model.ProjectWithExternalResult, error)
}

// commandExecutor は外部コマンド実行のインターフェース。
type commandExecutor interface {
	CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)
}

// defaultCommandExecutor は commandExecutor のデフォルト実装。
type defaultCommandExecutor struct{}

// CombinedOutput は指定コマンドを実行し、標準出力・標準エラー出力を結合して返します。
func (e *defaultCommandExecutor) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// forgejoClient は Forgejo API/CLI を用いた ExternalGitClient 実装。
type forgejoClient struct {
	exec   commandExecutor
	apiURL string
	token  string
	user   string
}

// newDefaultExternalGitClient は forgejoClient を生成します。
// 必要に応じて環境変数からForgejo API情報を取得
func newDefaultExternalGitClient() ExternalGitClient {
	apiURL := os.Getenv("FORGEJO_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:13031/api/v1"
	}
	token := os.Getenv("FORGEJO_TOKEN")
	user := os.Getenv("FORGEJO_USER")
	if user == "" {
		user = "musuhi"
	}
	return &forgejoClient{exec: &defaultCommandExecutor{}, apiURL: apiURL, token: token, user: user}
}

// CreateRepositoryWithExternal はForgejoリポジトリ作成と初回pushを実行します。
func (s *projectService) CreateRepositoryWithExternal(ctx context.Context, owner, repoName, visibility, localPath, commitMessage string) (*model.ProjectWithExternalResult, error) {
	owner = strings.TrimSpace(owner)
	repoName = strings.TrimSpace(repoName)
	visibility = strings.TrimSpace(strings.ToLower(visibility))
	localPath = strings.TrimSpace(localPath)
	commitMessage = strings.TrimSpace(commitMessage)

	if owner == "" {
		return nil, fmt.Errorf("%w: owner is required", ErrValidation)
	}
	if repoName == "" {
		return nil, fmt.Errorf("%w: repoName is required", ErrValidation)
	}
	if !projectNamePattern.MatchString(repoName) {
		return nil, fmt.Errorf("%w: repoName must match %s", ErrValidation, projectNamePattern.String())
	}
	if visibility != "public" && visibility != "private" {
		return nil, fmt.Errorf("%w: visibility must be public or private", ErrValidation)
	}
	if localPath == "" {
		return nil, fmt.Errorf("%w: localPath is required", ErrValidation)
	}
	if !filepath.IsAbs(localPath) {
		return nil, fmt.Errorf("%w: localPath must be absolute path", ErrValidation)
	}
	if commitMessage == "" {
		return nil, fmt.Errorf("%w: commitMessage is required", ErrValidation)
	}
	if len([]rune(commitMessage)) > 256 {
		return nil, fmt.Errorf("%w: commitMessage must be 256 characters or less", ErrValidation)
	}
	if stat, err := os.Stat(localPath); err != nil || !stat.IsDir() {
		return nil, fmt.Errorf("%w: localPath must be existing directory", ErrValidation)
	}

	result, err := s.gitClient.CreateRepositoryAndInitialPush(ctx, owner, repoName, visibility, localPath, commitMessage)
	if err != nil {
		return nil, fmt.Errorf("projectService.CreateRepositoryWithExternal: %w", err)
	}
	return result, nil
}

// isKnownGitHubInputError はGitHub CLIの入力エラーを判定します。
func isKnownGitHubInputError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "name already exists") ||
		strings.Contains(msg, "authentication failed") ||
		strings.Contains(msg, "not logged in")
}

// CreateRepositoryAndInitialPush はローカルリポジトリ初期化・Forgejoリポジトリ作成・初回pushを実行します。
func (c *forgejoClient) CreateRepositoryAndInitialPush(ctx context.Context, owner, repoName, visibility, localPath, commitMessage string) (*model.ProjectWithExternalResult, error) {
	if err := c.prepareLocalRepository(ctx, localPath, commitMessage); err != nil {
		return nil, err
	}

	// Forgejo APIでリポジトリ作成
	vis := "private"
	if visibility == "public" {
		vis = "public"
	}
	// curl -X POST -H "Authorization: token ..." -H "Content-Type: application/json" .../api/v1/user/repos -d '{...}'
	apiURL := c.apiURL + "/users/" + owner + "/repos"
	payload := fmt.Sprintf(`{"name":"%s","private":%v}`, repoName, vis == "private")
	args := []string{"-sS", "-X", "POST", "-H", "Authorization: token " + c.token, "-H", "Content-Type: application/json", apiURL, "-d", payload}
	out, err := c.exec.CombinedOutput(ctx, "curl", args...)
	if err != nil {
		return nil, fmt.Errorf("forgejo repo create failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	// origin URL
	repoURL := fmt.Sprintf("%s/%s/%s.git", strings.TrimSuffix(c.apiURL, "/api/v1"), owner, repoName)

	// git remote add & push
	remoteURL := fmt.Sprintf("http://%s:%s@%s/%s/%s.git", c.user, c.token, strings.TrimPrefix(strings.TrimPrefix(c.apiURL, "http://"), "https://"), owner, repoName)
	if out, err := c.exec.CombinedOutput(ctx, "git", "-C", localPath, "remote", "add", "origin", remoteURL); err != nil {
		msg := strings.ToLower(string(out))
		if !strings.Contains(msg, "remote origin already exists") {
			return nil, fmt.Errorf("git remote add origin failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}
	if out, err := c.exec.CombinedOutput(ctx, "git", "-C", localPath, "push", "-u", "origin", "main"); err != nil {
		return nil, fmt.Errorf("git push failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return &model.ProjectWithExternalResult{
		RepositoryURL:     repoURL,
		ExternalProjectID: repoName,
		PushStatus:        "success",
	}, nil
}

// prepareLocalRepository はローカルリポジトリの初期化・add・commit・mainブランチ作成・origin削除を行います。
func (c *forgejoClient) prepareLocalRepository(ctx context.Context, localPath, commitMessage string) error {
	if _, err := c.exec.CombinedOutput(ctx, "git", "-C", localPath, "rev-parse", "--is-inside-work-tree"); err != nil {
		out, initErr := c.exec.CombinedOutput(ctx, "git", "-C", localPath, "init")
		if initErr != nil {
			return fmt.Errorf("git init failed: %w: %s", initErr, strings.TrimSpace(string(out)))
		}
	}

	if out, err := c.exec.CombinedOutput(ctx, "git", "-C", localPath, "add", "."); err != nil {
		return fmt.Errorf("git add failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	commitArgs := []string{"-C", localPath, "-c", "user.name=Musuhi", "-c", "user.email=musuhi@example.com", "commit", "-m", commitMessage}
	if out, err := c.exec.CombinedOutput(ctx, "git", commitArgs...); err != nil {
		msg := strings.ToLower(string(out))
		if !strings.Contains(msg, "nothing to commit") && !strings.Contains(msg, "no changes added to commit") {
			return fmt.Errorf("git commit failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}

	if out, err := c.exec.CombinedOutput(ctx, "git", "-C", localPath, "branch", "-M", "main"); err != nil {
		return fmt.Errorf("git branch -M main failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	if out, err := c.exec.CombinedOutput(ctx, "git", "-C", localPath, "remote", "remove", "origin"); err != nil {
		msg := strings.ToLower(string(out))
		if !errors.Is(err, context.Canceled) &&
			!strings.Contains(msg, "no such remote") &&
			!strings.Contains(msg, "could not remove config section") {
			return fmt.Errorf("git remote remove origin failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}

	return nil
}
