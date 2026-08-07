package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
	if !strings.Contains(plainView, "macOS: brew install cloudflared") {
		t.Errorf("View missing macOS instruction, got:\n%s", plainView)
	}
	if !strings.Contains(plainView, "Windows: winget install --id Cloudflare.cloudflared") {
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
	if !strings.Contains(macLine, "brew install cloudflared") {
		t.Errorf("macOS command contains unexpected line wrap inside command: %q", macLine)
	}
	if !strings.Contains(winLine, "winget install --id Cloudflare.cloudflared") {
		t.Errorf("Windows command contains unexpected line wrap inside command: %q", winLine)
	}
	if !strings.Contains(linLine, "curl -L https://github.com/cloudflare") {
		t.Errorf("Linux command contains unexpected line wrap inside command: %q", linLine)
	}
}

func TestModelAllViews(t *testing.T) {
	m := NewModel(22, "/custom/path")

	// 1. statusCheckingPort
	m.status = statusCheckingPort
	v1 := stripANSI(m.View())
	if !strings.Contains(v1, "Checking if a service is listening on TCP port 22") {
		t.Errorf("statusCheckingPort view unexpected: %s", v1)
	}

	// 2. statusCheckingCloudflared (with warning)
	m.status = statusCheckingCloudflared
	m.portWarning = true
	v2 := stripANSI(m.View())
	if !strings.Contains(v2, "Warning: No active listener detected on port 22") {
		t.Errorf("statusCheckingCloudflared view unexpected: %s", v2)
	}

	// 3. statusDownloading
	m.status = statusDownloading
	m.downloadProgress = 0.45
	v3 := stripANSI(m.View())
	if !strings.Contains(v3, "Downloading cloudflared") {
		t.Errorf("statusDownloading view unexpected: %s", v3)
	}

	// 4. statusStartingTunnel with logs
	m.status = statusStartingTunnel
	m.logs = []string{"Log line 1", "Log line 2", "Log line 3", "Log line 4"}
	v4 := stripANSI(m.View())
	if !strings.Contains(v4, "Contacting trycloudflare.com") || !strings.Contains(v4, "Log line 4") {
		t.Errorf("statusStartingTunnel view unexpected: %s", v4)
	}

	// 5. statusError
	m.status = statusError
	m.err = fmt.Errorf("sample connection error")
	v5 := stripANSI(m.View())
	if !strings.Contains(v5, "Error occurred") || !strings.Contains(v5, "sample connection error") {
		t.Errorf("statusError view unexpected: %s", v5)
	}
}

func TestModelUpdates(t *testing.T) {
	// Test Init
	mInit := NewModel(8080, "")
	if cmd := mInit.Init(); cmd == nil {
		t.Errorf("Init() returned nil cmd")
	}

	// Test Key Msg 'q' -> quit
	mKey := NewModel(8080, "")
	_, cmd := mKey.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Errorf("Key 'q' did not return quit cmd")
	}

	// Test WindowSizeMsg
	mWin := NewModel(8080, "")
	updatedModel, _ := mWin.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if updatedModel.(Model).width != 100 || updatedModel.(Model).height != 40 {
		t.Errorf("WindowSizeMsg did not set dimensions correctly")
	}

	// Test portCheckMsg
	mPort := NewModel(8080, "")
	updatedModel, _ = mPort.Update(portCheckMsg{listening: false})
	m2 := updatedModel.(Model)
	if !m2.portWarning || m2.status != statusCheckingCloudflared {
		t.Errorf("portCheckMsg update failed: warning=%v status=%v", m2.portWarning, m2.status)
	}

	// Test cloudflaredCheckMsg success
	updatedModel, _ = m2.Update(cloudflaredCheckMsg{path: "/usr/local/bin/cloudflared"})
	m3 := updatedModel.(Model)
	if m3.cloudflaredPath != "/usr/local/bin/cloudflared" || m3.status != statusStartingTunnel {
		t.Errorf("cloudflaredCheckMsg update failed: path=%s status=%v", m3.cloudflaredPath, m3.status)
	}

	// Test tunnelUpdateMsg with URL
	updatedModel, _ = m3.Update(tunnelUpdateMsg{URL: "https://test.trycloudflare.com", Log: "Tunnel connected"})
	m4 := updatedModel.(Model)
	if m4.tunnelURL != "https://test.trycloudflare.com" || m4.status != statusTunnelRunning {
		t.Errorf("tunnelUpdateMsg URL update failed: url=%s status=%v", m4.tunnelURL, m4.status)
	}

	// Test tunnelUpdateMsg done with error
	updatedModel, _ = m4.Update(tunnelUpdateMsg{Done: true, Err: fmt.Errorf("network disconnect")})
	m5 := updatedModel.(Model)
	if m5.status != statusError || m5.err == nil {
		t.Errorf("tunnelUpdateMsg error update failed: status=%v err=%v", m5.status, m5.err)
	}
}

