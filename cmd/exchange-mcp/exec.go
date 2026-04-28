package main

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func execTool(args ...string) (*mcp.CallToolResult, any, error) {
	cmd := exec.Command("exchange", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := stderr.String()
		if msg == "" {
			msg = err.Error()
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: stdout.String()}},
	}, nil, nil
}

func checkCLI() error {
	_, err := exec.LookPath("exchange")
	if err != nil {
		return fmt.Errorf("exchange CLI not found in PATH. Install it first: https://github.com/dsswift/cli-exchange")
	}
	return nil
}
