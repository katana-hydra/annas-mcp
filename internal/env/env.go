package env

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/iosifache/annas-mcp/internal/logger"
	"github.com/iosifache/annas-mcp/internal/mirror"
	"go.uber.org/zap"
)

const DefaultAnnasBaseURL = "annas-archive.li"

type Env struct {
	SecretKey    string `json:"secret"`
	DownloadPath string `json:"download_path"`
	AnnasBaseURL string `json:"annas_base_url"`
}

var (
	resolvedEnvOnce        sync.Once
	resolvedAnnasBaseURL   string
	resolveAnnasBaseURLErr error
	resolveAnnasBaseURL    = defaultResolveAnnasBaseURL
)

func GetEnv() (*Env, error) {
	l := logger.GetLogger()

	secretKey := os.Getenv("ANNAS_SECRET_KEY")
	downloadPath := os.Getenv("ANNAS_DOWNLOAD_PATH")
	fixedBaseURL := normalizeBaseURL(os.Getenv("ANNAS_FIXED_BASE_URL"))
	fallbackBaseURL := normalizeBaseURL(os.Getenv("ANNAS_BASE_URL"))
	if secretKey == "" || downloadPath == "" {
		err := errors.New("ANNAS_SECRET_KEY and ANNAS_DOWNLOAD_PATH environment variables must be set")

		// Never log secret keys - use boolean flags instead
		l.Error("Environment variables not set",
			zap.Bool("ANNAS_SECRET_KEY_set", secretKey != ""),
			zap.String("ANNAS_DOWNLOAD_PATH", downloadPath),
			zap.String("ANNAS_FIXED_BASE_URL", fixedBaseURL),
			zap.String("ANNAS_BASE_URL", fallbackBaseURL),
			zap.Error(err),
		)

		return nil, err
	}

	if !filepath.IsAbs(downloadPath) {
		return nil, fmt.Errorf("ANNAS_DOWNLOAD_PATH must be an absolute path, got: %s", downloadPath)
	}

	annasBaseURL := fixedBaseURL
	if annasBaseURL == "" {
		resolvedEnvOnce.Do(func() {
			resolvedAnnasBaseURL, resolveAnnasBaseURLErr = resolveAnnasBaseURL()
		})

		if resolveAnnasBaseURLErr != nil {
			l.Warn("Automatic Anna mirror resolution failed, using fallback mirror",
				zap.String("ANNAS_BASE_URL", fallbackBaseURL),
				zap.Error(resolveAnnasBaseURLErr),
			)
		}

		annasBaseURL = normalizeBaseURL(resolvedAnnasBaseURL)
	}

	if annasBaseURL == "" {
		annasBaseURL = fallbackBaseURL
	}

	if annasBaseURL == "" {
		annasBaseURL = DefaultAnnasBaseURL
	}

	if fixedBaseURL != "" {
		l.Info("Using fixed Anna mirror from ANNAS_FIXED_BASE_URL",
			zap.String("ANNAS_FIXED_BASE_URL", fixedBaseURL),
		)
	}

	return &Env{
		SecretKey:    secretKey,
		DownloadPath: downloadPath,
		AnnasBaseURL: annasBaseURL,
	}, nil
}

func defaultResolveAnnasBaseURL() (string, error) {
	fallbackBaseURL := normalizeBaseURL(os.Getenv("ANNAS_BASE_URL"))
	resolver := mirror.NewResolver(nil, mirror.DefaultStatusPageURL, nil)
	return resolver.Resolve(context.Background(), mirror.ResolveOptions{FallbackBaseURL: fallbackBaseURL})
}

func normalizeBaseURL(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.TrimSuffix(value, "/")
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	return value
}

func resetResolvedEnvForTests() {
	resolvedEnvOnce = sync.Once{}
	resolvedAnnasBaseURL = ""
	resolveAnnasBaseURLErr = nil
}
