package tunnel

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
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
			goos:         "linux",
			goarch:       "386",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-386",
			expectedFile: "cloudflared",
		},
		{
			goos:         "linux",
			goarch:       "arm",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm",
			expectedFile: "cloudflared",
		},
		{
			goos:         "linux",
			goarch:       "mips",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64",
			expectedFile: "cloudflared",
		},
		{
			goos:         "darwin",
			goarch:       "amd64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-amd64.tgz",
			expectedFile: "cloudflared",
		},
		{
			goos:         "darwin",
			goarch:       "arm64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-arm64.tgz",
			expectedFile: "cloudflared",
		},
		{
			goos:         "darwin",
			goarch:       "ppc64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-darwin-amd64.tgz",
			expectedFile: "cloudflared",
		},
		{
			goos:         "windows",
			goarch:       "amd64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe",
			expectedFile: "cloudflared.exe",
		},
		{
			goos:         "windows",
			goarch:       "386",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-386.exe",
			expectedFile: "cloudflared.exe",
		},
		{
			goos:         "windows",
			goarch:       "arm64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe",
			expectedFile: "cloudflared.exe",
		},
		{
			goos:         "freebsd",
			goarch:       "amd64",
			expectedURL:  "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64",
			expectedFile: "cloudflared",
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
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create temporary listener: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	if !CheckPort(port) {
		t.Errorf("CheckPort() returned false for an active port %d, want true", port)
	}

	_ = listener.Close()

	if CheckPort(port) {
		t.Errorf("CheckPort() returned true for a closed port %d, want false", port)
	}
}

func TestParseTunnelURL(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantURL   string
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

func TestGetCloudflaredPath(t *testing.T) {
	// 1. Custom path that exists
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "custom-cloudflared")
	// #nosec G306
	if err := os.WriteFile(customPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("Failed to create custom binary: %v", err)
	}

	p, err := GetCloudflaredPath(customPath)
	if err != nil || p != customPath {
		t.Errorf("GetCloudflaredPath(existingCustom) = (%q, %v), want (%q, nil)", p, err, customPath)
	}

	// 2. Custom path that does not exist
	nonExistent := filepath.Join(tempDir, "nonexistent-cloudflared")
	_, err = GetCloudflaredPath(nonExistent)
	if err == nil {
		t.Errorf("GetCloudflaredPath(nonExistent) returned nil error, want error")
	}
}

func TestProgressReader(t *testing.T) {
	data := []byte("hello world 1234567890")
	var reportedProgress []float64

	pr := &progressReader{
		Reader: bytes.NewReader(data),
		Total:  int64(len(data)),
		OnProgress: func(p float64) {
			reportedProgress = append(reportedProgress, p)
		},
	}

	buf := make([]byte, 5)
	for {
		_, err := pr.Read(buf)
		if err != nil {
			break
		}
	}

	if len(reportedProgress) == 0 {
		t.Errorf("progressReader did not report any progress updates")
	}

	lastProgress := reportedProgress[len(reportedProgress)-1]
	if lastProgress != 1.0 {
		t.Errorf("Final progress = %v, want 1.0", lastProgress)
	}
}

func TestExtractTgzBinary(t *testing.T) {
	tempDir := t.TempDir()
	tgzPath := filepath.Join(tempDir, "test.tgz")
	destPath := filepath.Join(tempDir, "extracted-cloudflared")

	// Create valid tar.gz archive with a file named "cloudflared"
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	content := []byte("#!/bin/sh\necho cloudflared\n")
	hdr := &tar.Header{
		Name:     "cloudflared",
		Mode:     0755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("Failed writing tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Failed writing tar content: %v", err)
	}
	_ = tw.Close()
	_ = gw.Close()

	if err := os.WriteFile(tgzPath, buf.Bytes(), 0600); err != nil {
		t.Fatalf("Failed writing tgz file: %v", err)
	}

	// Extract
	if err := extractTgzBinary(tgzPath, destPath); err != nil {
		t.Fatalf("extractTgzBinary failed: %v", err)
	}

	// #nosec G304
	extracted, err := os.ReadFile(destPath)
	if err != nil || string(extracted) != string(content) {
		t.Errorf("Extracted content = %q, err = %v, want %q", string(extracted), err, string(content))
	}

	// Test archive without cloudflared binary
	var buf2 bytes.Buffer
	gw2 := gzip.NewWriter(&buf2)
	tw2 := tar.NewWriter(gw2)
	hdr2 := &tar.Header{
		Name:     "other-file",
		Mode:     0644,
		Size:     4,
		Typeflag: tar.TypeReg,
	}
	_ = tw2.WriteHeader(hdr2)
	_, _ = tw2.Write([]byte("test"))
	_ = tw2.Close()
	_ = gw2.Close()

	tgzPath2 := filepath.Join(tempDir, "invalid.tgz")
	_ = os.WriteFile(tgzPath2, buf2.Bytes(), 0600)
	err = extractTgzBinary(tgzPath2, destPath)
	if err == nil {
		t.Errorf("extractTgzBinary on archive without cloudflared returned nil error, want error")
	}
}

func TestRunTunnelInvalidBinary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := make(chan Update, 10)
	RunTunnel(ctx, "/path/to/nonexistent/binary", 8080, "", ch)

	var lastUpdate Update
	for u := range ch {
		lastUpdate = u
		if u.Done {
			break
		}
	}

	if lastUpdate.Err == nil {
		t.Errorf("RunTunnel with invalid binary returned no error")
	}
}

func TestRunTunnelInvalidBinaryWithDomain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := make(chan Update, 10)
	RunTunnel(ctx, "/path/to/nonexistent/binary", 8080, "example.com", ch)

	var lastUpdate Update
	for u := range ch {
		lastUpdate = u
		if u.Done {
			break
		}
	}

	if lastUpdate.Err == nil {
		t.Errorf("RunTunnel with invalid binary and domain returned no error")
	}
}
