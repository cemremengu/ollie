# Ollie

A CLI helper toolset for providing utility functions for Ollama.

## Installation

```bash
go build -o ollie
```

## Usage

```bash
./ollie --help
```

## Adding New Commands

To add a new command to Ollie:

1. Create a new file in the `cmd/` directory (e.g., `cmd/mycommand.go`)
2. Define your command using Cobra's command structure
3. Register the command with the root command in the `init()` function

### Example Command

Create a file `cmd/example.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "A brief description of your command",
	Long:  `A longer description that explains what your command does.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Example command executed!")
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)
}
```

That's it! Your new command will be automatically available when you rebuild the application.

## Project Structure

```
ollie/
├── go.mod              # Go module definition
├── main.go             # Application entry point
├── cmd/
│   └── root.go         # Root command definition
└── README.md           # This file
```

## Version

Current version: 0.1.0
