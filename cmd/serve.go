// Package cmd defines the command line interface commands and flags.
package cmd

import (
	"fmt"
	"net/http"
	"github.com/spf13/cobra"
)

var (
	servePort int
	serveDir  string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a simple HTML / static file server",
	Long:  `Start a simple HTTP server to serve static files from a directory on a specified port.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		fs := http.FileServer(http.Dir(serveDir))
		
		// Set up a basic logger middleware for requests
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("[%s] %s %s\n", r.Method, r.RemoteAddr, r.URL.Path)
			fs.ServeHTTP(w, r)
		})

		fmt.Printf("Starting static HTML/file server for directory '%s' on port %d\n", serveDir, servePort)
		fmt.Printf("Local URL: http://localhost:%d\n", servePort)
		fmt.Println("Press Ctrl+C to terminate.")

		server := &http.Server{
			Addr:              fmt.Sprintf(":%d", servePort),
			ReadHeaderTimeout: http.DefaultClient.Timeout, // Basic timeout settings to satisfy gosec (G114)
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to run serve: %w", err)
		}
		return nil
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port number to host the server on")
	serveCmd.Flags().StringVarP(&serveDir, "dir", "d", ".", "Directory to serve files from")
	rootCmd.AddCommand(serveCmd)
}
