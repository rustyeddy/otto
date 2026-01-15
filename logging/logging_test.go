package logging

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLevel_CaseInsensitive(t *testing.T) {
	cases := []struct {
		input string
		level string
	}{
		{input: "DEBUG", level: "DEBUG"},
		{input: "Info", level: "INFO"},
		{input: "warn", level: "WARN"},
		{input: "WARNING", level: "WARN"},
		{input: "error", level: "ERROR"},
	}

	for _, tc := range cases {
		level, err := ParseLevel(tc.input)
		require.NoError(t, err)
		assert.Equal(t, tc.level, level.String())
	}
}

func TestBuild_OutputString_ReturnsBuffer(t *testing.T) {
	cfg := Config{
		Level:  "info",
		Format: "text",
		Output: "string",
	}

	logger, closer, buf, err := Build(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
	assert.Nil(t, closer)
	require.NotNil(t, buf)

	logger.Info("hello")
	assert.NotEmpty(t, buf.String())
}

func TestBuild_OutputFile_ReturnsCloser(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "otto.log")
	cfg := Config{
		Level:    "info",
		Format:   "text",
		Output:   "file",
		FilePath: path,
	}

	logger, closer, buf, err := Build(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
	require.NotNil(t, closer)
	assert.Nil(t, buf)

	logger.Info("hello")
	require.NoError(t, closer.Close())

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotEmpty(t, contents)
}

func TestService_HTTP_GET_ReturnsConfig(t *testing.T) {
	svc, err := NewService(DefaultConfig())
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/log", nil)
	rec := httptest.NewRecorder()

	svc.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var cfg Config
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&cfg))
	assert.Equal(t, DefaultLevel, cfg.Level)
	assert.Equal(t, DefaultFormat, cfg.Format)
	assert.Equal(t, DefaultOutput, cfg.Output)
}

func TestService_HTTP_PUT_ValidConfig_Updates(t *testing.T) {
	svc, err := NewService(DefaultConfig())
	require.NoError(t, err)

	payload := Config{
		Level:  "DEBUG",
		Format: "json",
		Output: "string",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/log", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	svc.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var cfg Config
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&cfg))
	assert.Equal(t, "debug", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.Equal(t, "string", cfg.Output)

	current := svc.Config()
	assert.Equal(t, "debug", current.Level)
	assert.Equal(t, "json", current.Format)
	assert.Equal(t, "string", current.Output)
	assert.NotNil(t, current.Buffer)
}

func TestService_HTTP_PUT_InvalidConfig_400(t *testing.T) {
	svc, err := NewService(DefaultConfig())
	require.NoError(t, err)

	payload := Config{
		Level:  "info",
		Format: "text",
		Output: "file",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/log", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	svc.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.NotEmpty(t, resp.Error)
}
