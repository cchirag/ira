//go:build prod

package config

func init() {
	Current = new(Config)
	Current.Mode = "prod"
	Current.DaemonLogPath = getDaemonLogPath()
	Current.DaemonSocketPath = getDaemonSocketPath()
	Current.DaemonBinaryPath = getDaemonBinaryPath()
	Current.DaemonDBPath = getDaemonDBPath()
	Current.ClientLogPath = getClientLogPath()
	Current.ClientConfigPath = getClientConfigPath()
}
