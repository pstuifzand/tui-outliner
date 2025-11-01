package history

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Manager handles loading and saving history to TOML files
type Manager struct {
	historyDir string
}

// HistoryFile represents the structure of a history TOML file
type HistoryFile struct {
	Entries []string `toml:"entries"`
}

// NewManager creates a new history manager with directory at ~/.local/share/tui-outliner/history/
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	historyDir := filepath.Join(homeDir, ".local", "share", "tui-outliner", "history")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		historyDir: historyDir,
	}, nil
}

// Load loads history entries from a TOML file
func (m *Manager) Load(filename string) ([]string, error) {
	filePath := filepath.Join(m.historyDir, filename)

	// If file doesn't exist, return empty slice
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var histFile HistoryFile
	if err := toml.Unmarshal(data, &histFile); err != nil {
		// If parse error, return empty and continue (don't fail on corrupted file)
		return []string{}, nil
	}

	return histFile.Entries, nil
}

// Save saves history entries to a TOML file
func (m *Manager) Save(filename string, entries []string) error {
	filePath := filepath.Join(m.historyDir, filename)

	histFile := HistoryFile{
		Entries: entries,
	}

	data, err := toml.Marshal(histFile)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
