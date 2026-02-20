package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, ":8080", cfg.Proxy.ListenAddr)
	assert.Equal(t, 0.1, cfg.Proxy.SamplingRate)
	assert.Equal(t, 10000, cfg.Proxy.Buffer.QueueSize)
	assert.Equal(t, 20, cfg.Proxy.Buffer.Workers)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoad_FromFile(t *testing.T) {
	content := `
server:
  host: "127.0.0.1"
  port: 9090
log:
  level: "debug"
  format: "console"
`
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "tvc.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "console", cfg.Log.Format)
}

func TestLoad_InvalidFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("TVC_SERVER_PORT", "3000")
	t.Setenv("TVC_LOG_LEVEL", "error")

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, 3000, cfg.Server.Port)
	assert.Equal(t, "error", cfg.Log.Level)
}
