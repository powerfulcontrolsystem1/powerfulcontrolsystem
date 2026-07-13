package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type ToolSettings struct {
	Enabled        bool   `json:"enabled"`
	Command        string `json:"command"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type VulnerabilityToolSettings struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"`
	Command        string `json:"command"`
	TargetPath     string `json:"target_path"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type ScheduleSettings struct {
	Enabled bool   `json:"enabled"`
	Cron    string `json:"cron"`
}

type Settings struct {
	TargetHost         string                    `json:"target_host"`
	PortList           string                    `json:"port_list"`
	Profile            string                    `json:"profile"`
	MaxHistory         int                       `json:"max_history"`
	MaxFindingsPerTool int                       `json:"max_findings_per_tool"`
	DataDir            string                    `json:"data_dir"`
	ConfigPath         string                    `json:"config_path,omitempty"`
	Schedule           ScheduleSettings          `json:"schedule"`
	Lynis              ToolSettings              `json:"lynis"`
	Nmap               ToolSettings              `json:"nmap"`
	VulnerabilityScan  VulnerabilityToolSettings `json:"vulnerability_scan"`
}

type Manager struct {
	path string
	mu   sync.RWMutex
}

func ResolveBackendDir() string {
	if env := strings.TrimSpace(os.Getenv("PCS_BACKEND_DIR")); env != "" {
		return filepath.Clean(env)
	}
	workingDir, err := os.Getwd()
	if err == nil {
		for _, candidate := range []string{workingDir, filepath.Join(workingDir, "backend")} {
			if isBackendDir(candidate) {
				return filepath.Clean(candidate)
			}
		}
	}
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		if isBackendDir(execDir) {
			return filepath.Clean(execDir)
		}
	}
	return filepath.Clean(workingDir)
}

func isBackendDir(candidate string) bool {
	if strings.TrimSpace(candidate) == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(candidate, "go.mod")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(candidate, "handlers")); err != nil {
		return false
	}
	return true
}

func DefaultDataDir() string {
	return filepath.Join(ResolveBackendDir(), "logs", "vps_security")
}

func DefaultConfigPath() string {
	// The backend runs as the unprivileged `pcs` user in production. Keep the
	// mutable scanner configuration together with its private runtime data,
	// rather than below the application directory, which is read-only there.
	return filepath.Join(DefaultDataDir(), "config.json")
}

func legacyDefaultConfigPath() string {
	return filepath.Join(ResolveBackendDir(), "secure", "vps_security_config.json")
}

func DefaultSettings() Settings {
	dataDir := DefaultDataDir()
	settings := Settings{
		TargetHost:         "127.0.0.1",
		PortList:           "49222,80,443,5432,8080,8443",
		Profile:            "full",
		MaxHistory:         60,
		MaxFindingsPerTool: 150,
		DataDir:            dataDir,
		ConfigPath:         DefaultConfigPath(),
		Schedule: ScheduleSettings{
			Enabled: true,
			Cron:    "0 2 * * *",
		},
		Lynis: ToolSettings{
			Enabled:        runtime.GOOS == "linux",
			Command:        "lynis",
			TimeoutSeconds: 900,
		},
		Nmap: ToolSettings{
			Enabled:        runtime.GOOS == "linux",
			Command:        "nmap",
			TimeoutSeconds: 300,
		},
		VulnerabilityScan: VulnerabilityToolSettings{
			Enabled:        runtime.GOOS == "linux",
			Provider:       "trivy",
			Command:        "trivy",
			TargetPath:     "/",
			TimeoutSeconds: 1200,
		},
	}
	return settings
}

func NewManager(path string) *Manager {
	if strings.TrimSpace(path) == "" {
		path = DefaultConfigPath()
	}
	return &Manager{path: filepath.Clean(path)}
}

func (m *Manager) Path() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.path
}

func (m *Manager) Load() (Settings, error) {
	m.mu.RLock()
	path := m.path
	m.mu.RUnlock()
	settings := DefaultSettings()
	settings.ConfigPath = path
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if filepath.Clean(path) == filepath.Clean(DefaultConfigPath()) {
				if legacyRaw, legacyErr := os.ReadFile(legacyDefaultConfigPath()); legacyErr == nil {
					if jsonErr := json.Unmarshal(legacyRaw, &settings); jsonErr != nil {
						return settings, jsonErr
					}
					normalize(&settings)
					if saveErr := writeSettings(path, settings); saveErr != nil {
						return settings, saveErr
					}
					return settings, nil
				}
			}
			if saveErr := writeSettings(path, settings); saveErr != nil {
				return settings, saveErr
			}
			return settings, nil
		}
		return settings, err
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return settings, err
	}
	normalize(&settings)
	return settings, nil
}

func (m *Manager) Save(settings Settings) (Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings.ConfigPath = m.path
	normalize(&settings)
	if err := writeSettings(m.path, settings); err != nil {
		return settings, err
	}
	return settings, nil
}

func writeSettings(path string, settings Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o600)
}

func normalize(settings *Settings) {
	defaults := DefaultSettings()
	settings.TargetHost = fallback(settings.TargetHost, defaults.TargetHost)
	settings.PortList = fallback(settings.PortList, defaults.PortList)
	settings.Profile = strings.ToLower(strings.TrimSpace(fallback(settings.Profile, defaults.Profile)))
	if settings.Profile == "" {
		settings.Profile = defaults.Profile
	}
	if settings.MaxHistory <= 0 {
		settings.MaxHistory = defaults.MaxHistory
	}
	if settings.MaxFindingsPerTool <= 0 {
		settings.MaxFindingsPerTool = defaults.MaxFindingsPerTool
	}
	settings.DataDir = fallback(settings.DataDir, defaults.DataDir)
	settings.Schedule.Cron = fallback(settings.Schedule.Cron, defaults.Schedule.Cron)
	settings.Lynis.Command = fallback(settings.Lynis.Command, defaults.Lynis.Command)
	settings.Nmap.Command = fallback(settings.Nmap.Command, defaults.Nmap.Command)
	settings.VulnerabilityScan.Command = fallback(settings.VulnerabilityScan.Command, defaults.VulnerabilityScan.Command)
	settings.VulnerabilityScan.Provider = strings.ToLower(strings.TrimSpace(fallback(settings.VulnerabilityScan.Provider, defaults.VulnerabilityScan.Provider)))
	settings.VulnerabilityScan.TargetPath = fallback(settings.VulnerabilityScan.TargetPath, defaults.VulnerabilityScan.TargetPath)
	if settings.Lynis.TimeoutSeconds <= 0 {
		settings.Lynis.TimeoutSeconds = defaults.Lynis.TimeoutSeconds
	}
	if settings.Nmap.TimeoutSeconds <= 0 {
		settings.Nmap.TimeoutSeconds = defaults.Nmap.TimeoutSeconds
	}
	if settings.VulnerabilityScan.TimeoutSeconds <= 0 {
		settings.VulnerabilityScan.TimeoutSeconds = defaults.VulnerabilityScan.TimeoutSeconds
	}
	if settings.ConfigPath == "" {
		settings.ConfigPath = DefaultConfigPath()
	}
	settings.TargetHost = strings.TrimSpace(settings.TargetHost)
	settings.PortList = strings.TrimSpace(settings.PortList)
	settings.DataDir = filepath.Clean(settings.DataDir)
	settings.ConfigPath = filepath.Clean(settings.ConfigPath)
	settings.Schedule.Cron = strings.TrimSpace(settings.Schedule.Cron)
	settings.Lynis.Command = strings.TrimSpace(settings.Lynis.Command)
	settings.Nmap.Command = strings.TrimSpace(settings.Nmap.Command)
	settings.VulnerabilityScan.Command = strings.TrimSpace(settings.VulnerabilityScan.Command)
	settings.VulnerabilityScan.TargetPath = strings.TrimSpace(settings.VulnerabilityScan.TargetPath)
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
