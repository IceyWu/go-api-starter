package banner

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
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

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visibleRuneLen(s string) int {
	return len([]rune(ansiRegexp.ReplaceAllString(s, "")))
}

func padRightVisible(s string, width int) string {
	pad := width - visibleRuneLen(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

func bannerStyle() string {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("BANNER_STYLE"))); v != "" {
		switch v {
		case "ascii", "unicode":
			return v
		}
	}

	// Default: prefer unicode in modern terminals.
	if runtime.GOOS == "windows" {
		if os.Getenv("WT_SESSION") != "" {
			return "unicode"
		}
		// Older consoles can render poorly; keep ASCII as a safer fallback.
		return "ascii"
	}
	return "unicode"
}

type bannerBox struct {
	topLeft     string
	topRight    string
	bottomLeft  string
	bottomRight string
	horizontal  string
	vertical    string
	sepLeft     string
	sepRight    string
}

func (b bannerBox) top(width int) string {
	return b.topLeft + strings.Repeat(b.horizontal, width) + b.topRight
}

func (b bannerBox) bottom(width int) string {
	return b.bottomLeft + strings.Repeat(b.horizontal, width) + b.bottomRight
}

func (b bannerBox) sep(width int) string {
	return b.sepLeft + strings.Repeat(b.horizontal, width) + b.sepRight
}

func (b bannerBox) line(content string, width int) string {
	return b.vertical + padRightVisible(content, width) + b.vertical
}

// PrintBanner prints the startup banner
func PrintBanner(appName, env, port, localIP string) {
	useColor := isColorSupported()
	style := bannerStyle()

	var bc, reset, green, yellow, magenta string
	if useColor {
		bc = Bold + Cyan
		reset = Reset
		green = Bold + Green
		yellow = Yellow
		magenta = Magenta
	}

	// Compose content lines (include left padding inside the box).
	arrow := "➤"
	if style == "ascii" {
		arrow = ">"
	}

	lineTitle := "  " + green + "[*] " + appName + " started successfully!" + reset
	lineEnv := "  " + yellow + arrow + " Environment:" + reset + "  " + env
	lineLocal := "  " + green + arrow + " Local:" + reset + "        http://localhost:" + port
	lineNetwork := "  " + green + arrow + " Network:" + reset + "      http://" + localIP + ":" + port
	lineAPIBase := "  " + magenta + arrow + " API Base:" + reset + "     http://localhost:" + port + "/api/v1"
	lineDocs := "  " + magenta + arrow + " API Docs:" + reset + "     http://localhost:" + port + "/docs"
	lineSwagger := "  " + magenta + arrow + " Swagger:" + reset + "      http://localhost:" + port + "/swagger/index.html"
	lineOpenAPI := "  " + magenta + arrow + " OpenAPI:" + reset + "      http://localhost:" + port + "/swagger/doc.json"

	sections := [][]string{
		{lineTitle},
		{lineEnv},
		{lineLocal, lineNetwork},
		{lineAPIBase, lineDocs, lineSwagger, lineOpenAPI},
	}

	// Determine inner width based on visible characters.
	innerWidth := 0
	for _, sec := range sections {
		for _, l := range sec {
			if n := visibleRuneLen(l); n > innerWidth {
				innerWidth = n
			}
		}
	}
	// Add one trailing space so the right border isn't tight.
	innerWidth += 1

	var box bannerBox
	if style == "unicode" {
		box = bannerBox{
			topLeft:     "╔",
			topRight:    "╗",
			bottomLeft:  "╚",
			bottomRight: "╝",
			horizontal:  "═",
			vertical:    "║",
			sepLeft:     "╠",
			sepRight:    "╣",
		}
	} else {
		box = bannerBox{
			topLeft:     "+",
			topRight:    "+",
			bottomLeft:  "+",
			bottomRight: "+",
			horizontal:  "-",
			vertical:    "|",
			sepLeft:     "+",
			sepRight:    "+",
		}
	}

	fmt.Println()
	fmt.Println(bc + box.top(innerWidth) + reset)
	for si, sec := range sections {
		for _, l := range sec {
			fmt.Println(bc + box.line(l, innerWidth) + reset)
		}
		if si != len(sections)-1 {
			fmt.Println(bc + box.sep(innerWidth) + reset)
		}
	}
	fmt.Println(bc + box.bottom(innerWidth) + reset)
	fmt.Println()
}
