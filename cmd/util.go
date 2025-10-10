package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// getOllamaModelsPath returns the path to the Ollama models directory.
// It first checks the OLLAMA_MODELS environment variable.
// If not set, it defaults to ~/.ollama/models.
func getOllamaModelsPath() (string, error) {
	modelPath := os.Getenv("OLLAMA_MODELS")
	if modelPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		modelPath = filepath.Join(home, ".ollama", "models")
	}
	return modelPath, nil
}
