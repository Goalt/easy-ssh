package tunnel

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// CheckPort checks if any service is listening on the specified port.
func CheckPort(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetCloudflaredDownloadURL returns the download URL and the expected binary filename for cloudflared.
func GetCloudflaredDownloadURL(goos, goarch string) (string, string) {
	filename := "cloudflared"
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}

	var platformArch string
	switch goos {
	case "linux":
		switch goarch {
		case "amd64":
			platformArch = "linux-amd64"
		case "arm64":
			platformArch = "linux-arm64"
		case "386":
			platformArch = "linux-386"
		case "arm":
			platformArch = "linux-arm"
		default:
			platformArch = "linux-amd64"
		}
	case "darwin":
		switch goarch {
		case "amd64":
			platformArch = "darwin-amd64"
		case "arm64":
			platformArch = "darwin-arm64"
		default:
			platformArch = "darwin-amd64"
		}
	case "windows":
		switch goarch {
		case "amd64":
			platformArch = "windows-amd64"
		case "386":
			platformArch = "windows-386"
		default:
			platformArch = "windows-amd64"
		}
	default:
		platformArch = "linux-amd64"
	}

	// Darwin assets are packaged as .tgz archives
	if goos == "darwin" {
		url := fmt.Sprintf("https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-%s.tgz", platformArch)
		return url, filename + ext
	}

	url := fmt.Sprintf("https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-%s%s", platformArch, ext)
	return url, filename + ext
}

// GetCloudflaredPath searches for cloudflared in the PATH or the local cache folder.
func GetCloudflaredPath(customPath string) (string, error) {
	if customPath != "" {
		if _, err := os.Stat(customPath); err != nil {
			return "", fmt.Errorf("custom cloudflared path not found: %w", err)
		}
		return customPath, nil
	}

	// 1. Check system PATH
	if path, err := exec.LookPath("cloudflared"); err == nil {
		return path, nil
	}

	// 2. Check local cache directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".easy-ssh", "bin")
	_, expectedName := GetCloudflaredDownloadURL(runtime.GOOS, runtime.GOARCH)
	cachePath := filepath.Join(cacheDir, expectedName)

	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	return "", fmt.Errorf("cloudflared not found in PATH or cache")
}

// DownloadCloudflared downloads cloudflared for the current OS/Arch.
func DownloadCloudflared(onProgress func(float64)) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".easy-ssh", "bin")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	url, expectedName := GetCloudflaredDownloadURL(runtime.GOOS, runtime.GOARCH)
	destPath := filepath.Join(cacheDir, expectedName)

	// Create request
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download cloudflared: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download cloudflared, HTTP status: %s", resp.Status)
	}

	// Support progress reporting
	totalSize := resp.ContentLength
	reader := &progressReader{
		Reader:    resp.Body,
		Total:     totalSize,
		OnProgress: onProgress,
	}

	if runtime.GOOS == "darwin" {
		// Extract from .tgz
		tempTgzPath := destPath + ".tgz"
		tempTgzFile, err := os.OpenFile(tempTgzPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to create temporary tgz file: %w", err)
		}
		defer os.Remove(tempTgzPath)

		if _, err := io.Copy(tempTgzFile, reader); err != nil {
			tempTgzFile.Close()
			return "", fmt.Errorf("failed to write temporary tgz file: %w", err)
		}
		tempTgzFile.Close()

		// Extract the binary from the tgz
		if err := extractTgzBinary(tempTgzPath, destPath); err != nil {
			return "", fmt.Errorf("failed to extract tgz archive: %w", err)
		}
	} else {
		// Write direct binary
		out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create executable: %w", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, reader); err != nil {
			return "", fmt.Errorf("failed to save executable: %w", err)
		}
	}

	// Make sure it is executable (especially on unix systems)
	if err := os.Chmod(destPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make cloudflared executable: %w", err)
	}

	return destPath, nil
}

type progressReader struct {
	io.Reader
	Total     int64
	ReadBytes int64
	OnProgress func(float64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		pr.ReadBytes += int64(n)
		if pr.Total > 0 && pr.OnProgress != nil {
			pr.OnProgress(float64(pr.ReadBytes) / float64(pr.Total))
		}
	}
	return n, err
}

func extractTgzBinary(tarGzPath, destPath string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for the "cloudflared" binary
		if header.Typeflag == tar.TypeReg && (header.Name == "cloudflared" || filepath.Base(header.Name) == "cloudflared") {
			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("cloudflared binary not found inside tgz archive")
}

var TryCloudflareRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)

// ParseTunnelURL parses a single line of log and returns the tunnel URL if found.
func ParseTunnelURL(line string) (string, bool) {
	match := TryCloudflareRegex.FindString(line)
	if match != "" {
		return match, true
	}
	return "", false
}

// TunnelUpdate represents an update sent from the tunnel background runner.
type TunnelUpdate struct {
	URL  string
	Log  string
	Err  error
	Done bool
}

// RunTunnel executes cloudflared in a subprocess, parsing its logs for the trycloudflare URL.
func RunTunnel(ctx context.Context, binPath string, port int, ch chan<- TunnelUpdate) {
	cmd := exec.CommandContext(ctx, binPath, "tunnel", "--url", fmt.Sprintf("tcp://127.0.0.1:%d", port))

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	// Graceful shutdown on context cancellation
	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			// Try sending SIGINT (os.Interrupt) for clean shutdown
			_ = cmd.Process.Signal(os.Interrupt)
			time.AfterFunc(3*time.Second, func() {
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
			})
		}
	}()

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		ch <- TunnelUpdate{Err: fmt.Errorf("failed to start cloudflared: %w", err), Done: true}
		return
	}

	// Scan stdout/stderr merged output line by line
	go func() {
		defer pw.Close()
		defer pr.Close()

		var r io.Reader = pr
		buf := make([]byte, 4096)
		var leftover string

		for {
			n, err := r.Read(buf)
			if n > 0 {
				chunk := leftover + string(buf[:n])
				lines := strings.Split(chunk, "\n")
				if len(lines) > 0 {
					leftover = lines[len(lines)-1]
					for _, line := range lines[:len(lines)-1] {
						trimmed := strings.TrimSpace(line)
						if trimmed == "" {
							continue
						}

						// Scan for tunnel URL
						if url, found := ParseTunnelURL(trimmed); found {
							ch <- TunnelUpdate{URL: url, Log: trimmed}
						} else {
							ch <- TunnelUpdate{Log: trimmed}
						}
					}
				}
			}
			if err != nil {
				break
			}
		}
	}()

	err := cmd.Wait()
	ch <- TunnelUpdate{Err: err, Done: true}
}

