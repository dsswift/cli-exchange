package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--setup":
			if err := setupClaude(); err != nil {
				fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
				os.Exit(1)
			}
			return
		case "--version":
			fmt.Println(Version)
			return
		case "--help", "-h":
			fmt.Println("exchange-mcp - MCP server for the Exchange CLI")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  exchange-mcp          Start the MCP server (stdio)")
			fmt.Println("  exchange-mcp --setup  Register in Claude Code settings")
			fmt.Println("  exchange-mcp --version")
			return
		}
	}

	if err := checkCLI(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "exchange",
		Version: Version,
	}, nil)

	registerTools(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func claudeConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude.json")
}

func setupClaude() error {
	configPath := claudeConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("~/.claude.json not found. Install Claude Code first, then re-run --setup.")
			return nil
		}
		return fmt.Errorf("reading %s: %w", configPath, err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parsing %s: %w", configPath, err)
	}

	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		servers = make(map[string]any)
		config["mcpServers"] = servers
	}

	// Check existing entry
	binName := "exchange-mcp"
	if runtime.GOOS == "windows" {
		binName = "exchange-mcp.exe"
	}

	if existing, ok := servers["exchange"].(map[string]any); ok {
		cmd, _ := existing["command"].(string)
		if cmd == binName {
			fmt.Println("exchange MCP server already configured in ~/.claude.json")
			return nil
		}
	}

	// Preserve env vars from old config
	var envVars map[string]any
	if existing, ok := servers["exchange"].(map[string]any); ok {
		if env, ok := existing["env"].(map[string]any); ok {
			envVars = make(map[string]any)
			for _, key := range []string{"EXCHANGE_CLIENT_ID", "EXCHANGE_TENANT_ID"} {
				if v, ok := env[key]; ok {
					envVars[key] = v
				}
			}
		}
	}

	entry := map[string]any{
		"type":    "stdio",
		"command": binName,
	}
	if len(envVars) > 0 {
		entry["env"] = envVars
	}

	servers["exchange"] = entry

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Atomic write
	tmpFile := configPath + ".tmp"
	if err := os.WriteFile(tmpFile, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmpFile, configPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("replacing config: %w", err)
	}

	fmt.Println("Registered exchange MCP server in ~/.claude.json")
	return nil
}
