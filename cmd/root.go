package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/Goalt/easy-ssh/pkg/ui"
)

var (
	port           int
	customCFPath   string
)

var rootCmd = &cobra.Command{
	Use:   "easy-ssh",
	Short: "easy-ssh starts a Cloudflare tunnel on a specified port (default 22)",
	Long: `easy-ssh is a CLI tool that simplifies creating Cloudflare tunnels for TCP services
like SSH. It checks if a service is listening on the target port, automatically downloads
and runs cloudflared if needed, and presents a beautiful, modern terminal interface.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create and run the Bubble Tea program
		m := ui.NewModel(port, customCFPath)
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run UI: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&port, "port", "p", 22, "Port number to tunnel")
	rootCmd.Flags().StringVarP(&customCFPath, "cloudflared-path", "c", "", "Path to custom cloudflared binary (optional)")
}
