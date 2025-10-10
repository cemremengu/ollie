package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
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

// getOllamaUIDGID looks up the ollama user and group and returns their UID and GID.
// If the ollama user or group is not found, it returns -1 for both (indicating no chown should occur).
func getOllamaUIDGID() (int, int, error) {
	ollamaUser, err := user.Lookup("ollama")
	if err != nil {
		// If ollama user doesn't exist, don't change ownership
		return -1, -1, nil
	}

	ollamaGroup, err := user.LookupGroup("ollama")
	if err != nil {
		// If ollama group doesn't exist, don't change ownership
		return -1, -1, nil
	}

	uid, err := strconv.Atoi(ollamaUser.Uid)
	if err != nil {
		return -1, -1, fmt.Errorf("failed to parse UID: %w", err)
	}

	gid, err := strconv.Atoi(ollamaGroup.Gid)
	if err != nil {
		return -1, -1, fmt.Errorf("failed to parse GID: %w", err)
	}

	return uid, gid, nil
}
