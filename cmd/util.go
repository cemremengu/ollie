package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

const systemPath = "/usr/share/ollama/.ollama/models"

// getOllamaModelsPath returns the path to the Ollama models directory.
// It first checks the OLLAMA_MODELS environment variable.
// If not set, it defaults to /usr/share/ollama/.ollama, falling back to ~/.ollama/models if not found.
func getOllamaModelsPath() (string, error) {
	modelPath := os.Getenv("OLLAMA_MODELS")
	if modelPath == "" {
		if _, err := os.Stat(systemPath); err == nil {
			modelPath = systemPath
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			modelPath = filepath.Join(home, ".ollama", "models")
		}
	}
	return modelPath, nil
}
