package cmd

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ulikunitz/xz"
)

// extractTarball extracts a tarball to the specified destination directory
func extractTarball(fileName, destPath string) error {
	// Get ollama user/group ownership
	uid, gid, err := getOllamaUIDGID()
	if err != nil {
		return fmt.Errorf("failed to get ollama user/group: %w", err)
	}
	// Open the tarball file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create the appropriate reader based on file extension
	var tarReader *tar.Reader

	if strings.HasSuffix(fileName, ".tar.xz") {
		xzReader, err := xz.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create xz reader: %w", err)
		}
		tarReader = tar.NewReader(xzReader)
	} else if strings.HasSuffix(fileName, ".tar.gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else if strings.HasSuffix(fileName, ".tar.bz2") || strings.HasSuffix(fileName, ".tar.bz") {
		bzReader := bzip2.NewReader(file)
		tarReader = tar.NewReader(bzReader)
	} else if strings.HasSuffix(fileName, ".tar") {
		tarReader = tar.NewReader(file)
	} else {
		return fmt.Errorf("unsupported file extension for %s", fileName)
	}

	// Extract files from the tarball
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct full path
		targetPath := filepath.Join(destPath, header.Name)

		// Handle directory entries
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			// Set ownership to ollama:ollama if user/group exists
			if uid != -1 && gid != -1 {
				if err := os.Chown(targetPath, uid, gid); err != nil {
					return fmt.Errorf("failed to set ownership for directory %s: %w", targetPath, err)
				}
			}
			continue
		}

		// Create parent directories for files
		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
		}
		// Set ownership on parent directory
		if uid != -1 && gid != -1 {
			if err := os.Chown(parentDir, uid, gid); err != nil {
				return fmt.Errorf("failed to set ownership for parent directory %s: %w", parentDir, err)
			}
		}

		// Create and write file
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}

		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}
		outFile.Close()

		// Set ownership on the file
		if uid != -1 && gid != -1 {
			if err := os.Chown(targetPath, uid, gid); err != nil {
				return fmt.Errorf("failed to set ownership for file %s: %w", targetPath, err)
			}
		}
	}

	return nil
}

var loadCmd = &cobra.Command{
	Use:   "load TARBALL_FILE",
	Short: "Load an Ollama model from a tarball",
	Long: `Load an Ollama model by extracting a tarball to the Ollama models directory.
Supports .tar, .tar.gz, .tar.bz/.tar.bz2, and .tar.xz formats.

The tarball is extracted to the directory specified by the OLLAMA_MODELS
environment variable, or ~/.ollama/models if not set.

Examples:
  ollie load llama2.tar
  ollie load llama2.tar.gz
  ollie load llama2.tar.xz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName := args[0]

		// Get model path from environment or use default
		modelPath, err := getOllamaModelsPath()
		if err != nil {
			return err
		}

		// Extract tarball
		if err := extractTarball(fileName, modelPath); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Successfully loaded model from %s to %s\n", fileName, modelPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}
