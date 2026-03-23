package env

import (
	"sync/atomic"
	"testing"
)

func TestGetEnvUsesFixedBaseURLWithoutResolver(t *testing.T) {
	t.Setenv("ANNAS_SECRET_KEY", "secret")
	t.Setenv("ANNAS_DOWNLOAD_PATH", t.TempDir())
	t.Setenv("ANNAS_FIXED_BASE_URL", "fixed.example")
	t.Setenv("ANNAS_BASE_URL", "fallback.example")

	resetResolvedEnvForTests()

	resolverCalls := int32(0)
	resolveAnnasBaseURL = func() (string, error) {
		atomic.AddInt32(&resolverCalls, 1)
		return "dynamic.example", nil
	}
	t.Cleanup(func() {
		resolveAnnasBaseURL = defaultResolveAnnasBaseURL
		resetResolvedEnvForTests()
	})

	cfg, err := GetEnv()
	if err != nil {
		t.Fatalf("GetEnv returned error: %v", err)
	}

	if cfg.AnnasBaseURL != "fixed.example" {
		t.Fatalf("expected fixed.example, got %q", cfg.AnnasBaseURL)
	}

	if atomic.LoadInt32(&resolverCalls) != 0 {
		t.Fatalf("expected resolver not to be called, got %d calls", resolverCalls)
	}
}

func TestGetEnvCachesResolvedBaseURL(t *testing.T) {
	t.Setenv("ANNAS_SECRET_KEY", "secret")
	t.Setenv("ANNAS_DOWNLOAD_PATH", t.TempDir())
	t.Setenv("ANNAS_BASE_URL", "fallback.example")

	resetResolvedEnvForTests()

	resolverCalls := int32(0)
	resolveAnnasBaseURL = func() (string, error) {
		atomic.AddInt32(&resolverCalls, 1)
		return "dynamic.example", nil
	}
	t.Cleanup(func() {
		resolveAnnasBaseURL = defaultResolveAnnasBaseURL
		resetResolvedEnvForTests()
	})

	first, err := GetEnv()
	if err != nil {
		t.Fatalf("first GetEnv returned error: %v", err)
	}

	second, err := GetEnv()
	if err != nil {
		t.Fatalf("second GetEnv returned error: %v", err)
	}

	if first.AnnasBaseURL != "dynamic.example" || second.AnnasBaseURL != "dynamic.example" {
		t.Fatalf("expected cached dynamic.example, got %q and %q", first.AnnasBaseURL, second.AnnasBaseURL)
	}

	if atomic.LoadInt32(&resolverCalls) != 1 {
		t.Fatalf("expected resolver to be called once, got %d calls", resolverCalls)
	}
}

func TestGetEnvFallsBackWhenResolverFails(t *testing.T) {
	t.Setenv("ANNAS_SECRET_KEY", "secret")
	t.Setenv("ANNAS_DOWNLOAD_PATH", t.TempDir())
	t.Setenv("ANNAS_BASE_URL", "fallback.example")

	resetResolvedEnvForTests()

	resolveAnnasBaseURL = func() (string, error) {
		return "", assertiveError("resolver failed")
	}
	t.Cleanup(func() {
		resolveAnnasBaseURL = defaultResolveAnnasBaseURL
		resetResolvedEnvForTests()
	})

	cfg, err := GetEnv()
	if err != nil {
		t.Fatalf("GetEnv returned error: %v", err)
	}

	if cfg.AnnasBaseURL != "fallback.example" {
		t.Fatalf("expected fallback.example, got %q", cfg.AnnasBaseURL)
	}
}

type assertiveError string

func (e assertiveError) Error() string { return string(e) }
