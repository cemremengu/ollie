package cmd

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ModelName represents the parsed components of an Ollama model name
type ModelName struct {
	Host      string
	Namespace string
	Model     string
	Tag       string
}

// Manifest represents the structure of an Ollama manifest file
type Manifest struct {
	Config struct {
		Digest string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		Digest string `json:"digest"`
	} `json:"layers"`
}

// parseModelName parses an Ollama model name into its components
// Supports formats:
//   - host/namespace/model:tag
//   - namespace/model:tag
//   - namespace/model
//   - model:tag
//   - model
func parseModelName(name string) (*ModelName, error) {
	patterns := []string{
		`^(?P<host>[^/]+)/(?P<namespace>[^/]+)/(?P<model>[^:]+):(?P<tag>.+)$`, // host/namespace/model:tag
		`^(?P<namespace>[^/]+)/(?P<model>[^:]+):(?P<tag>.+)$`,                 // namespace/model:tag
		`^(?P<namespace>[^/]+)/(?P<model>[^:]+)$`,                             // namespace/model
		`^(?P<model>[^:]+):(?P<tag>.+)$`,                                      // model:tag
		`^(?P<model>[^:]+)$`,                                                  // model
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(name); matches != nil {
			result := &ModelName{
				Host:      "registry.ollama.ai",
				Namespace: "library",
				Model:     "",
				Tag:       "latest",
			}

			// Extract named groups
			names := re.SubexpNames()
			for i, match := range matches {
				switch names[i] {
				case "host":
					result.Host = match
				case "namespace":
					result.Namespace = match
				case "model":
					result.Model = match
				case "tag":
					result.Tag = match
				}
			}

			return result, nil
		}
	}

	return nil, fmt.Errorf("invalid model name format: %s", name)
}

// parseManifest reads and parses the manifest file, returning blob SHAs
func parseManifest(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	shas := []string{}

	// Add config digest
	if manifest.Config.Digest != "" {
		shas = append(shas, "sha256-"+strings.TrimPrefix(manifest.Config.Digest, "sha256:"))
	}

	// Add layer digests
	for _, layer := range manifest.Layers {
		if layer.Digest != "" {
			shas = append(shas, "sha256-"+strings.TrimPrefix(layer.Digest, "sha256:"))
		}
	}

	return shas, nil
}

// getFilePaths returns the relative paths for the manifest and all blobs
func getFilePaths(modelName *ModelName, modelPath string) ([]string, error) {
	manifestPath := filepath.Join(
		modelPath,
		"manifests",
		modelName.Host,
		modelName.Namespace,
		modelName.Model,
		modelName.Tag,
	)

	blobShas, err := parseManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	paths := []string{}

	// Add manifest path (relative)
	paths = append(paths, filepath.Join(
		"manifests",
		modelName.Host,
		modelName.Namespace,
		modelName.Model,
		modelName.Tag,
	))

	// Add blob paths
	for _, sha := range blobShas {
		paths = append(paths, filepath.Join("blobs", sha))
	}

	return paths, nil
}

// createTarball creates a tarball from the given paths and writes to stdout
func createTarball(modelPath string, relativePaths []string) error {
	tw := tar.NewWriter(os.Stdout)
	defer tw.Close()

	for _, relPath := range relativePaths {
		absPath := filepath.Join(modelPath, relPath)

		// Get file info
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", absPath, err)
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create header for %s: %w", relPath, err)
		}

		// Use relative path in tarball
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header for %s: %w", relPath, err)
		}

		// Write file content
		file, err := os.Open(absPath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", absPath, err)
		}

		if _, err := io.Copy(tw, file); err != nil {
			file.Close()
			return fmt.Errorf("failed to write %s to tarball: %w", relPath, err)
		}
		file.Close()
	}

	return nil
}

var saveCmd = &cobra.Command{
	Use:   "save MODEL_NAME",
	Short: "Save an Ollama model to a tarball",
	Long: `Save an Ollama model by creating a tarball containing its manifest and blob files.
The tarball is written to stdout, so you can redirect it to a file or pipe it elsewhere.

Examples:
  ollie save llama2 > llama2.tar
  ollie save library/llama2:latest > llama2.tar
  ollie save registry.ollama.ai/library/llama2:latest > llama2.tar`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelNameStr := args[0]

		// Check if stdout is a terminal
		if term.IsTerminal(int(os.Stdout.Fd())) {
			return fmt.Errorf("refusing to write binary tarball to terminal\nPlease redirect output to a file: ollie save %s > output.tar", modelNameStr)
		}

		// Parse model name
		modelName, err := parseModelName(modelNameStr)
		if err != nil {
			return err
		}

		// Get model path from environment or use default
		modelPath, err := getOllamaModelsPath()
		if err != nil {
			return err
		}

		// Get file paths
		filePaths, err := getFilePaths(modelName, modelPath)
		if err != nil {
			return err
		}

		// Create tarball
		if err := createTarball(modelPath, filePaths); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
