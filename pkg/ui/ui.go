package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Goalt/easy-ssh/pkg/tunnel"
)

type SessionStatus int

const (
	StatusCheckingPort SessionStatus = iota
	StatusCheckingCloudflared
	StatusDownloading
	StatusStartingTunnel
	StatusTunnelRunning
	StatusError
)

type progressUpdate struct {
	progress float64
	done     bool
	path     string
	err      error
}

type portCheckMsg struct {
	listening bool
}

type cloudflaredCheckMsg struct {
	path string
	err  error
}

type tunnelUpdateMsg tunnel.TunnelUpdate

// Model represents the state of the easy-ssh TUI.
type Model struct {
	status             SessionStatus
	port               int
	customPath         string
	portWarning        bool
	cloudflaredPath    string
	downloadProgress   float64
	tunnelURL          string
	logs               []string
	err                error
	spinner            spinner.Model
	progressBar        progress.Model
	ctx                context.Context
	cancel             context.CancelFunc
	chDownloadProgress chan progressUpdate
	chTunnelUpdate     chan tunnel.TunnelUpdate
	width              int
	height             int
}

// NewModel creates a new easy-ssh UI model.
func NewModel(port int, customPath string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(PrimaryColor)

	prog := progress.New(progress.WithDefaultGradient())

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		status:             StatusCheckingPort,
		port:               port,
		customPath:         customPath,
		spinner:            s,
		progressBar:        prog,
		ctx:                ctx,
		cancel:             cancel,
		chDownloadProgress: make(chan progressUpdate, 20),
		chTunnelUpdate:     make(chan tunnel.TunnelUpdate, 200),
	}
}

// Init initializes the Bubble Tea application.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.checkPortCmd(),
	)
}

// checkPortCmd performs the port check.
func (m Model) checkPortCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(300 * time.Millisecond) // Short pause for a smoother UI transition
		listening := tunnel.CheckPort(m.port)
		return portCheckMsg{listening: listening}
	}
}

// checkCloudflaredCmd checks if cloudflared is available.
func (m Model) checkCloudflaredCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(300 * time.Millisecond) // Short pause for a smoother UI transition
		path, err := tunnel.GetCloudflaredPath(m.customPath)
		return cloudflaredCheckMsg{path: path, err: err}
	}
}

// downloadCloudflaredCmd starts the cloudflared binary download.
func (m Model) downloadCloudflaredCmd() tea.Cmd {
	return func() tea.Msg {
		go func() {
			path, err := tunnel.DownloadCloudflared(func(p float64) {
				m.chDownloadProgress <- progressUpdate{progress: p}
			})
			m.chDownloadProgress <- progressUpdate{done: true, path: path, err: err}
		}()
		return m.listenToDownload()()
	}
}

// listenToDownload listens for updates on the download progress channel.
func (m Model) listenToDownload() tea.Cmd {
	return func() tea.Msg {
		update := <-m.chDownloadProgress
		return update
	}
}

// startTunnelCmd starts the cloudflared tunnel subprocess.
func (m Model) startTunnelCmd() tea.Cmd {
	return func() tea.Msg {
		go tunnel.RunTunnel(m.ctx, m.cloudflaredPath, m.port, m.chTunnelUpdate)
		return m.listenToTunnel()()
	}
}

// listenToTunnel listens for updates on the tunnel execution channel.
func (m Model) listenToTunnel() tea.Cmd {
	return func() tea.Msg {
		update := <-m.chTunnelUpdate
		return tunnelUpdateMsg(update)
	}
}

// Update handles state changes in response to messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progressBar.Width = msg.Width - 10
		if m.progressBar.Width > 60 {
			m.progressBar.Width = 60
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case portCheckMsg:
		if !msg.listening {
			m.portWarning = true
		}
		m.status = StatusCheckingCloudflared
		cmds = append(cmds, m.checkCloudflaredCmd())

	case cloudflaredCheckMsg:
		if msg.err != nil {
			m.status = StatusDownloading
			cmds = append(cmds, m.downloadCloudflaredCmd())
		} else {
			m.cloudflaredPath = msg.path
			m.status = StatusStartingTunnel
			cmds = append(cmds, m.startTunnelCmd())
		}

	case progressUpdate:
		if msg.done {
			if msg.err != nil {
				m.err = msg.err
				m.status = StatusError
			} else {
				m.cloudflaredPath = msg.path
				m.status = StatusStartingTunnel
				cmds = append(cmds, m.startTunnelCmd())
			}
		} else {
			m.downloadProgress = msg.progress
			cmds = append(cmds, m.listenToDownload())
		}

	case tunnelUpdateMsg:
		if msg.Done {
			// Check if the exit was due to our own context cancellation
			if m.ctx.Err() != nil {
				// Clean exit on interrupt/cancellation
				return m, tea.Quit
			}

			if msg.Err != nil {
				m.err = msg.Err
				m.status = StatusError
			} else if m.status != StatusTunnelRunning {
				// Tunnel exited before we found the URL
				m.err = fmt.Errorf("tunnel closed unexpectedly")
				m.status = StatusError
			} else {
				// Clean exit without error
				return m, tea.Quit
			}
		} else {
			if msg.URL != "" {
				m.tunnelURL = msg.URL
				m.status = StatusTunnelRunning
			}
			if msg.Log != "" {
				m.logs = append(m.logs, msg.Log)
				if len(m.logs) > 50 {
					m.logs = m.logs[1:] // Cap logs at last 50 lines
				}
			}
			cmds = append(cmds, m.listenToTunnel())
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the terminal user interface.
func (m Model) View() string {
	var sb strings.Builder

	// Header Banner
	sb.WriteString("\n")
	sb.WriteString(TitleStyle.Render(" EASY-SSH CLOUDFLARE TUNNEL ") + "\n")
	sb.WriteString(SubtitleStyle.Render(" Securely share your local SSH or TCP services ") + "\n\n")

	// Main body depending on status
	switch m.status {
	case StatusCheckingPort:
		sb.WriteString(fmt.Sprintf("%s Checking if a service is listening on TCP port %d...\n", m.spinner.View(), m.port))

	case StatusCheckingCloudflared:
		sb.WriteString(renderCheckedItem("Port checked successfully.", true))
		if m.portWarning {
			sb.WriteString(WarningBoxStyle.Render(fmt.Sprintf("⚠️  Warning: No active listener detected on port %d.\nGenerating tunnel anyway.", m.port)) + "\n\n")
		} else {
			sb.WriteString(TickStyle.Render("✔") + fmt.Sprintf(" Active listener detected on port %d.\n\n", m.port))
		}
		sb.WriteString(fmt.Sprintf("%s Checking for cloudflared installation...\n", m.spinner.View()))

	case StatusDownloading:
		sb.WriteString(renderCheckedItem(fmt.Sprintf("Port checked successfully (Port %d).", m.port), true))
		if m.portWarning {
			sb.WriteString(WarningBoxStyle.Render(fmt.Sprintf("⚠️  Warning: No active listener detected on port %d.\nGenerating tunnel anyway.", m.port)) + "\n\n")
		}
		sb.WriteString(renderCheckedItem("cloudflared not found in PATH.", false))
		sb.WriteString("Downloading cloudflared from GitHub Releases...\n")
		sb.WriteString(m.progressBar.ViewAs(m.downloadProgress) + fmt.Sprintf(" %.0f%%\n", m.downloadProgress*100))

	case StatusStartingTunnel:
		sb.WriteString(renderCheckedItem(fmt.Sprintf("Port checked successfully (Port %d).", m.port), true))
		if m.portWarning {
			sb.WriteString(WarningBoxStyle.Render(fmt.Sprintf("⚠️  Warning: No active listener detected on port %d.\nGenerating tunnel anyway.", m.port)) + "\n\n")
		}
		sb.WriteString(renderCheckedItem("cloudflared is ready.", true))
		sb.WriteString(fmt.Sprintf("%s Contacting trycloudflare.com to establish tunnel...\n", m.spinner.View()))
		if len(m.logs) > 0 {
			sb.WriteString("\n" + StatusStyle.Render("Logs:") + "\n")
			lastLogs := m.logs
			if len(lastLogs) > 3 {
				lastLogs = lastLogs[len(lastLogs)-3:]
			}
			for _, log := range lastLogs {
				sb.WriteString(StatusStyle.Render("  "+log) + "\n")
			}
		}

	case StatusTunnelRunning:
		if m.portWarning {
			sb.WriteString(WarningBoxStyle.Render(fmt.Sprintf("⚠️  Warning: No active listener detected on port %d.\nGenerating tunnel anyway.", m.port)) + "\n\n")
		}

		// Success Card
		var content strings.Builder
		content.WriteString(TickStyle.Render("✔ SUCCESS! Your Cloudflare Tunnel is established and active.") + "\n\n")
		content.WriteString(LabelStyle.Render("Tunnel URL:") + " " + CommandStyle.Render(m.tunnelURL) + "\n\n")
		content.WriteString("To connect via SSH from a remote client, use this command:\n")
		content.WriteString(CommandStyle.Render(fmt.Sprintf("ssh -o \"ProxyCommand=cloudflared access tcp --hostname %%h --port %%p\" user@%s", strings.TrimPrefix(m.tunnelURL, "https://"))) + "\n\n")
		content.WriteString(StatusStyle.Render("Note: The connecting client must also have the `cloudflared` binary installed."))

		sb.WriteString(SuccessBoxStyle.Render(content.String()) + "\n\n")
		sb.WriteString(m.spinner.View() + " " + StatusStyle.Render("Tunneling traffic... Press ") + CommandStyle.Render("Ctrl+C") + StatusStyle.Render(" or ") + CommandStyle.Render("Q") + StatusStyle.Render(" to terminate.") + "\n")

	case StatusError:
		sb.WriteString(CrossStyle.Render("✖ Error occurred:") + "\n")
		sb.WriteString(WarningBoxStyle.Render(m.err.Error()) + "\n\n")
		if len(m.logs) > 0 {
			sb.WriteString(StatusStyle.Render("Last logs:") + "\n")
			lastLogs := m.logs
			if len(lastLogs) > 5 {
				lastLogs = lastLogs[len(lastLogs)-5:]
			}
			for _, log := range lastLogs {
				sb.WriteString(StatusStyle.Render("  "+log) + "\n")
			}
			sb.WriteString("\n")
		}
		sb.WriteString(StatusStyle.Render("Press Ctrl+C or Q to exit.") + "\n")
	}

	return sb.String()
}

func renderCheckedItem(text string, success bool) string {
	if success {
		return fmt.Sprintf(" %s %s\n", TickStyle.Render("✔"), text)
	}
	return fmt.Sprintf(" %s %s\n", CrossStyle.Render("✖"), text)
}
