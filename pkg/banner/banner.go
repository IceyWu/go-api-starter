package banner

import (
	"fmt"
	"os"
	"runtime"
)

// ANSI color codes
const (
	Reset   = "\033[0m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Bold    = "\033[1m"
)

// isColorSupported checks if terminal supports colors
func isColorSupported() bool {
	// Disable colors in non-interactive or production environments
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("APP_ENV") == "production" {
		return false
	}
	// Check if running in a terminal
	if runtime.GOOS == "windows" {
		return os.Getenv("TERM") != "" || os.Getenv("WT_SESSION") != ""
	}
	// Linux/Mac: check if stdout is a terminal
	fi, _ := os.Stdout.Stat()
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// PrintBanner prints the startup banner
func PrintBanner(appName, env, port, localIP string) {
	useColor := isColorSupported()

	var bc, reset, green, yellow, magenta string
	if useColor {
		bc = Bold + Cyan
		reset = Reset
		green = Bold + Green
		yellow = Yellow
		magenta = Magenta
	}

	fmt.Println()
	fmt.Println(bc + "+================================================================+" + reset)
	fmt.Println(bc + "|" + reset + "  " + green + "[*] " + appName + " started successfully!" + reset + "                   " + bc + "|" + reset)
	fmt.Println(bc + "+================================================================+" + reset)
	fmt.Println(bc + "|" + reset + "  " + yellow + "> Environment:" + reset + "  " + env + "                                      " + bc + "|" + reset)
	fmt.Println(bc + "+----------------------------------------------------------------+" + reset)
	fmt.Println(bc + "|" + reset + "  " + green + "> Local:" + reset + "        http://localhost:" + port + "                         " + bc + "|" + reset)
	fmt.Println(bc + "|" + reset + "  " + green + "> Network:" + reset + "      http://" + localIP + ":" + port + "                      " + bc + "|" + reset)
	fmt.Println(bc + "+----------------------------------------------------------------+" + reset)
	fmt.Println(bc + "|" + reset + "  " + magenta + "> API Base:" + reset + "     http://localhost:" + port + "/api/v1                  " + bc + "|" + reset)
	fmt.Println(bc + "|" + reset + "  " + magenta + "> API Docs:" + reset + "     http://localhost:" + port + "/docs                     " + bc + "|" + reset)
	fmt.Println(bc + "|" + reset + "  " + magenta + "> Swagger:" + reset + "      http://localhost:" + port + "/swagger/index.html      " + bc + "|" + reset)
	fmt.Println(bc + "|" + reset + "  " + magenta + "> OpenAPI:" + reset + "      http://localhost:" + port + "/swagger/doc.json        " + bc + "|" + reset)
	fmt.Println(bc + "+================================================================+" + reset)
	fmt.Println()
}
