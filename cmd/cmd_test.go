package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmdFlags(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd --help returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "easy-ssh is a CLI tool") {
		t.Errorf("rootCmd help output missing description, got:\n%s", out)
	}
}

func TestServeCmdFlags(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"serve", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("serveCmd --help returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Start a simple HTTP server") {
		t.Errorf("serveCmd help output missing description, got:\n%s", out)
	}
}
