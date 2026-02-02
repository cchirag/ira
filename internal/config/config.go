package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Mode             string
	DaemonLogPath    string
	DaemonSocketPath string
	DaemonBinaryPath string
	DaemonDBPath     string
	ClientLogPath    string
	ClientConfigPath string
	// Add other common config fields here
}

var Current *Config

func getMacOSAppSupportDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		cacheDir, _ := os.UserCacheDir()
		if cacheDir == "" {
			cacheDir = os.TempDir()
		}
		return filepath.Join(cacheDir, "ira")
	}

	appSupport := filepath.Join(dir, "Library", "Application Support")
	testPath := filepath.Join(appSupport, "ira")

	if err := os.MkdirAll(testPath, 0755); err != nil {
		cacheDir, _ := os.UserCacheDir()
		if cacheDir == "" {
			cacheDir = os.TempDir()
		}
		return filepath.Join(cacheDir, "ira")
	}

	return testPath
}

func getAbsDevPath(relPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return relPath
	}
	return filepath.Join(cwd, relPath)
}

func getDaemonLogPath() string {
	var logDir string
	switch Current.Mode {
	case "dev":
		logDir = getAbsDevPath(filepath.Join("tmp", "log"))
	case "prod":
		if runtime.GOOS == "darwin" {
			dir, err := os.UserHomeDir()
			if err != nil {
				dir = os.TempDir()
			}
			logDir = filepath.Join(dir, "Library", "Logs", "ira")
		} else {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				cacheDir = os.TempDir()
			}
			logDir = filepath.Join(cacheDir, "ira", "log")
		}
	default:
		panic(fmt.Sprintf("invalid mode %q - allowed: dev, prod", Current.Mode))
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to create log directory %s: %v\n", logDir, err)
	}

	return filepath.Join(logDir, "daemon.log")
}

func getDaemonSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\irad`
	}

	var runtimeDir string
	switch Current.Mode {
	case "dev":
		runtimeDir = getAbsDevPath(filepath.Join("tmp", "run"))
	case "prod":
		runtimeDir = os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			runtimeDir = filepath.Join(os.TempDir(), fmt.Sprintf("ira-%d", os.Getuid()))
		}
	default:
		panic(fmt.Sprintf("invalid mode %q - allowed: dev, prod", Current.Mode))
	}

	if err := os.MkdirAll(runtimeDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to create runtime directory %s: %v\n", runtimeDir, err)
	}

	return filepath.Join(runtimeDir, "irad.sock")
}

func getDaemonBinaryPath() string {
	var cacheDir string

	switch Current.Mode {
	case "dev":
		cacheDir = getAbsDevPath(filepath.Join("tmp", "bin"))
	case "prod":
		dir, err := os.UserCacheDir()
		if err != nil {
			dir = os.TempDir()
		}
		cacheDir = filepath.Join(dir, "ira", "bin")
	default:
		panic(fmt.Sprintf("invalid mode %q - allowed: dev, prod", Current.Mode))
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to create cache directory %s: %v\n", cacheDir, err)
	}

	binaryName := "irad"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	return filepath.Join(cacheDir, binaryName)
}

func getDaemonDBPath() string {
	var dataDir string

	switch Current.Mode {
	case "dev":
		dataDir = getAbsDevPath(filepath.Join("tmp", "data"))
	case "prod":
		if runtime.GOOS == "darwin" {
			dataDir = getMacOSAppSupportDir()
		} else {
			dir, err := os.UserConfigDir()
			if err != nil {
				dir, err = os.UserHomeDir()
				if err != nil {
					dir = os.TempDir()
				}
				dir = filepath.Join(dir, ".ira")
			} else {
				dir = filepath.Join(dir, "ira")
			}
			dataDir = dir
		}
	default:
		panic(fmt.Sprintf("invalid mode %q - allowed: dev, prod", Current.Mode))
	}

	if Current.Mode != "prod" || runtime.GOOS != "darwin" {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to create data directory %s: %v\n", dataDir, err)
		}
	}

	return filepath.Join(dataDir, "daemon.db")
}

func getClientLogPath() string {
	var logDir string

	switch Current.Mode {
	case "dev":
		logDir = getAbsDevPath(filepath.Join("tmp", "log"))
	case "prod":
		if runtime.GOOS == "darwin" {
			dir, err := os.UserHomeDir()
			if err != nil {
				dir = os.TempDir()
			}
			logDir = filepath.Join(dir, "Library", "Logs", "ira")
		} else {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				cacheDir = os.TempDir()
			}
			logDir = filepath.Join(cacheDir, "ira", "log")
		}
	default:
		panic(fmt.Sprintf("invalid mode %q - allowed: dev, prod", Current.Mode))
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to create log directory %s: %v\n", logDir, err)
	}

	return filepath.Join(logDir, "client.log")
}

func getClientConfigPath() string {
	var configDir string

	switch Current.Mode {
	case "dev":
		configDir = getAbsDevPath(filepath.Join("tmp", "config"))
	case "prod":
		if runtime.GOOS == "darwin" {
			configDir = getMacOSAppSupportDir()
		} else {
			dir, err := os.UserConfigDir()
			if err != nil {
				dir, err = os.UserHomeDir()
				if err != nil {
					dir = os.TempDir()
				}
				dir = filepath.Join(dir, ".ira")
			} else {
				dir = filepath.Join(dir, "ira")
			}
			configDir = dir
		}
	default:
		panic(fmt.Sprintf("invalid mode %q - allowed: dev, prod", Current.Mode))
	}

	if Current.Mode != "prod" || runtime.GOOS != "darwin" {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to create config directory %s: %v\n", configDir, err)
		}
	}

	return filepath.Join(configDir, "config.yml")
}
