//go:build darwin

package focus

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type cmdProvider struct {
	binaryPath string
	logger     *slog.Logger
}

type cmdOutput struct {
	TsMs        int64  `json:"ts_ms"`
	AppName     string `json:"app_name"`
	BundleID    string `json:"bundle_id"`
	PID         int    `json:"pid"`
	WindowTitle string `json:"window_title"`
}

func newProvider(logger *slog.Logger) (provider, error) {
	binaryPath, err := ensureFocusBinary(logger)
	if err != nil {
		return nil, err
	}
	return &cmdProvider{binaryPath: binaryPath, logger: logger}, nil
}

func (c *cmdProvider) Current() (FocusSnapshot, error) {
	cmd := exec.Command(c.binaryPath)
	output, err := cmd.Output()
	if err != nil {
		return FocusSnapshot{}, fmt.Errorf("focusd failed: %w", err)
	}
	var parsed cmdOutput
	if err := json.Unmarshal(output, &parsed); err != nil {
		return FocusSnapshot{}, fmt.Errorf("decode focusd output: %w", err)
	}
	return FocusSnapshot{
		TsMs:        parsed.TsMs,
		AppName:     parsed.AppName,
		BundleID:    parsed.BundleID,
		PID:         parsed.PID,
		WindowTitle: parsed.WindowTitle,
	}, nil
}

func ensureFocusBinary(logger *slog.Logger) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", ".."))
	sourcePath := filepath.Join(repoRoot, "cmd", "focusd", "main.swift")
	if _, err := os.Stat(sourcePath); err != nil {
		return "", fmt.Errorf("focusd source missing: %w", err)
	}
	binDir := filepath.Join(repoRoot, "services", "core-go", "bin")
	binaryPath := filepath.Join(binDir, "focusd")
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", fmt.Errorf("create focusd bin dir: %w", err)
	}

	cmd := exec.Command("xcrun", "swiftc", "-O", sourcePath, "-o", binaryPath)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("focusd build failed", slog.String("output", string(output)))
		return "", fmt.Errorf("build focusd: %w", err)
	}
	return binaryPath, nil
}
