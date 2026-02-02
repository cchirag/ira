// spawn package provides methods to spawn the "irad" daemon as a seperate process.
// We run the daemon in a detached mode, to prevent the server from terminating when a user closes the ira (client) process

package spawn

import (
	"embed"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/cchirag/ira/internal/config"
)

// Run the daemon process as a seperate process
func RunDaemon(fs embed.FS) error {
	if isDaemonRunning() {
		return nil
	}

	name := binaryName()

	if err := copyBinary(name, fs); err != nil {
		return err
	}

	cmd := exec.Command(config.Current.DaemonBinaryPath)

	if err := spawn(cmd); err != nil {
		return fmt.Errorf("failed to spawn daemon: %w", err)
	}

	// Wait for daemon to be ready with retries
	maxRetries := 10
	retryDelay := 200 * time.Millisecond

	for range maxRetries {
		if isDaemonRunning() {
			return nil
		}
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("daemon failed to start after %d retries", maxRetries)
}

func isDaemonRunning() bool {
	conn, err := net.DialTimeout("unix", config.Current.DaemonSocketPath, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func copyBinary(name string, fs embed.FS) error {
	data, err := fs.ReadFile(fmt.Sprintf("bin/%s", name))
	if err != nil {
		return err
	}

	if err := os.WriteFile(config.Current.DaemonBinaryPath, data, 0755); err != nil {
		return err
	}

	return nil
}
