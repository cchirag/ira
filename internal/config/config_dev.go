//go:build dev

package config

func init() {
	Current = new(Config)
	Current.Mode = "dev"
	Current.DaemonLogPath = getDaemonLogPath()
	Current.DaemonSocketPath = getDaemonSocketPath()
	Current.DaemonBinaryPath = getDaemonBinaryPath()
	Current.DaemonDBPath = getDaemonDBPath()
	Current.ClientLogPath = getClientLogPath()
	Current.ClientConfigPath = getClientConfigPath()
}
