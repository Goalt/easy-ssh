package ui

import (
	"regexp"
	"strings"
	"testing"
)

// stripANSI removes ANSI escape sequences from a string.
func stripANSI(str string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansi.ReplaceAllString(str, "")
}

func TestModelTunnelRunningView(t *testing.T) {
	// Create a model and set state to statusTunnelRunning
	m := NewModel(8080, "")
	m.status = statusTunnelRunning
	m.tunnelURL = "https://test-tunnel.trycloudflare.com"

	// Get rendered view
	view := m.View()
	plainView := stripANSI(view)

	// Assertions for installation instructions
	if !strings.Contains(plainView, "macOS:") || !strings.Contains(plainView, "brew install cloudflared") {
		t.Errorf("View missing macOS instruction, got:\n%s", plainView)
	}
	if !strings.Contains(plainView, "Windows:") || !strings.Contains(plainView, "winget install --id Cloudflare.cloudflared") {
		t.Errorf("View missing Windows instruction, got:\n%s", plainView)
	}
	if !strings.Contains(plainView, "Linux (amd64):") {
		t.Errorf("View missing Linux instruction, got:\n%s", plainView)
	}

	// Verify alignment of instructions: they should start back-to-back without leading space offsets
	lines := strings.Split(plainView, "\n")
	var macLine, winLine, linLine string
	for _, l := range lines {
		lTrimmed := strings.TrimSpace(l)
		if strings.HasPrefix(lTrimmed, "macOS:") {
			macLine = l
		} else if strings.HasPrefix(lTrimmed, "Windows:") {
			winLine = l
		} else if strings.HasPrefix(lTrimmed, "Linux (amd64):") {
			linLine = l
		}
	}

	if macLine == "" {
		t.Fatal("macOS instruction line not found in view")
	}
	if winLine == "" {
		t.Fatal("Windows instruction line not found in view")
	}
	if linLine == "" {
		t.Fatal("Linux instruction line not found in view")
	}

	// Check if macOS, Windows, and Linux are left-aligned/start with same indentation.
	// Specifically, they should have no leading spaces inside the content blocks.
	if strings.HasPrefix(macLine, " ") || strings.HasPrefix(macLine, "\t") {
		t.Errorf("macOS instruction has unexpected leading indentation/whitespace: %q", macLine)
	}
	if strings.HasPrefix(winLine, " ") || strings.HasPrefix(winLine, "\t") {
		t.Errorf("Windows instruction has unexpected leading indentation/whitespace: %q", winLine)
	}
	if strings.HasPrefix(linLine, " ") || strings.HasPrefix(linLine, "\t") {
		t.Errorf("Linux instruction has unexpected leading indentation/whitespace: %q", linLine)
	}

	// Also verify that the copyable commands do not have any newlines inside them
	// We check for the presence of the commands themselves in the plainView output.
	if !strings.Contains(plainView, "brew install cloudflared") {
		t.Errorf("plainView missing macOS install command")
	}
	if !strings.Contains(plainView, "winget install --id Cloudflare.cloudflared") {
		t.Errorf("plainView missing Windows install command")
	}
	if !strings.Contains(plainView, "curl -L https://github.com/cloudflare") {
		t.Errorf("plainView missing Linux install command")
	}
}
