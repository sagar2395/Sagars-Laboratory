package cmd

import (
	"fmt"
	"io/fs"
	osExec "os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/sagars-lab/labctl/internal/api"
	"github.com/sagars-lab/labctl/ui"
)

var uiPort string

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the web UI dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := ":" + uiPort
		url := fmt.Sprintf("http://localhost:%s", uiPort)

		fmt.Printf("Starting labctl web UI at %s\n", url)
		fmt.Println("Press Ctrl+C to stop.")

		// Try to open browser
		go openBrowser(url)

		// Use embedded UI assets (sub-directory "dist" within the embed.FS)
		uiFS, _ := fs.Sub(ui.DistFS, "dist")
		server := api.NewServer(cfg, exec, reg, scenes, svcReg, uiFS)
		return server.Start(addr)
	},
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = osExec.Command("xdg-open", url).Start()
	case "darwin":
		err = osExec.Command("open", url).Start()
	case "windows":
		err = osExec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	if err != nil {
		fmt.Printf("Could not open browser: %v\n", err)
	}
}

func init() {
	uiCmd.Flags().StringVar(&uiPort, "port", "3939", "port to serve the UI on")
	rootCmd.AddCommand(uiCmd)
}
