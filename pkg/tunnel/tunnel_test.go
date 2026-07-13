package tunnel

import (
	"fmt"
	"net"
	"testing"
)

func TestGetCloudflaredDownloadURL(t *testing.T) {
	tests := []struct {
		goos         string
		goarch       string
		expectedURL  string
		expectedFile string
	}{
		{
			goos:         "linux",
			goarch:       "amd64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64",
			expectedFile: "cloudflared",
		},
		{
			goos:         "linux",
			goarch:       "arm64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64",
			expectedFile: "cloudflared",
		},
		{
			goos:         "darwin",
			goarch:       "amd64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-amd64.tgz",
			expectedFile: "cloudflared",
		},
		{
			goos:         "windows",
			goarch:       "amd64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe",
			expectedFile: "cloudflared.exe",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.goos, tt.goarch), func(t *testing.T) {
			url, file := GetCloudflaredDownloadURL(tt.goos, tt.goarch)
			if url != tt.expectedURL {
				t.Errorf("GetCloudflaredDownloadURL() url = %v, want %v", url, tt.expectedURL)
			}
			if file != tt.expectedFile {
				t.Errorf("GetCloudflaredDownloadURL() file = %v, want %v", file, tt.expectedFile)
			}
		})
	}
}

func TestCheckPort(t *testing.T) {
	// First check a port that should be closed (we'll use a random unassigned high port)
	// We dynamically find an open port, then close it to ensure nothing is listening on it.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create temporary listener: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// It's currently listening, so CheckPort should return true
	if !CheckPort(port) {
		t.Errorf("CheckPort() returned false for an active port %d, want true", port)
	}

	// Close it
	listener.Close()

	// Now it's closed, so CheckPort should return false
	if CheckPort(port) {
		t.Errorf("CheckPort() returned true for a closed port %d, want false", port)
	}
}

func TestParseTunnelURL(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantURL  string
		wantFound bool
	}{
		{
			name:      "simple trycloudflare url",
			line:      "https://example.trycloudflare.com",
			wantURL:   "https://example.trycloudflare.com",
			wantFound: true,
		},
		{
			name:      "boxed trycloudflare url",
			line:      "| https://some-subdomain-123.trycloudflare.com     |",
			wantURL:   "https://some-subdomain-123.trycloudflare.com",
			wantFound: true,
		},
		{
			name:      "no trycloudflare url",
			line:      "2026-07-13T19:59:46Z INF Requesting new quick Tunnel on trycloudflare.com...",
			wantURL:   "",
			wantFound: false,
		},
		{
			name:      "unrelated https url",
			line:      "Visit https://github.com/Goalt/easy-ssh for more info",
			wantURL:   "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotFound := ParseTunnelURL(tt.line)
			if gotFound != tt.wantFound {
				t.Errorf("ParseTunnelURL() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
			if gotURL != tt.wantURL {
				t.Errorf("ParseTunnelURL() gotURL = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}
