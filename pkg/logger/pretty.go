package logger

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	
	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	
	// Bright foreground colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
	
	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// PrettyLogger provides formatted logging methods
type PrettyLogger struct {
	prefix string
}

// NewPrettyLogger creates a new pretty logger
func NewPrettyLogger(prefix string) *PrettyLogger {
	return &PrettyLogger{prefix: prefix}
}

// Header prints a section header
func (l *PrettyLogger) Header(title string) {
	line := strings.Repeat("─", 80)
	fmt.Printf("\n%s%s%s%s\n", Bold, Cyan, line, Reset)
	fmt.Printf("%s%s▶ %s%s\n", Bold, BrightCyan, title, Reset)
	fmt.Printf("%s%s%s%s\n", Bold, Cyan, line, Reset)
}

// Footer prints a section footer
func (l *PrettyLogger) Footer(title string) {
	line := strings.Repeat("─", 80)
	fmt.Printf("%s%s%s%s\n", Bold, Cyan, line, Reset)
	fmt.Printf("%s%s✓ %s%s\n", Bold, BrightGreen, title, Reset)
	fmt.Printf("%s%s%s%s\n\n", Bold, Cyan, line, Reset)
}

// Info prints an info message
func (l *PrettyLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[INFO]%s %s\n", BrightBlue, Reset, msg)
}

// Success prints a success message
func (l *PrettyLogger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[✓]%s %s%s%s\n", BrightGreen, Reset, Green, msg, Reset)
}

// Warning prints a warning message
func (l *PrettyLogger) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[!]%s %s%s%s\n", BrightYellow, Reset, Yellow, msg, Reset)
}

// Error prints an error message
func (l *PrettyLogger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[✗]%s %s%s%s\n", BrightRed, Reset, Red, msg, Reset)
}

// Step prints a step message with number
func (l *PrettyLogger) Step(step int, total int, title string) {
	fmt.Printf("\n%s%s[%d/%d]%s %s%s%s\n", Bold, BrightMagenta, step, total, Reset, Bold, title, Reset)
}

// Detail prints a detail message (indented)
func (l *PrettyLogger) Detail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s│%s %s\n", BrightBlack, Reset, msg)
}

// DetailSuccess prints a success detail (indented)
func (l *PrettyLogger) DetailSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s│%s %s✓%s %s\n", BrightBlack, Reset, BrightGreen, Reset, msg)
}

// DetailError prints an error detail (indented)
func (l *PrettyLogger) DetailError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s│%s %s✗%s %s\n", BrightBlack, Reset, BrightRed, Reset, msg)
}

// DetailWarning prints a warning detail (indented)
func (l *PrettyLogger) DetailWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s│%s %s!%s %s\n", BrightBlack, Reset, BrightYellow, Reset, msg)
}

// KeyValue prints a key-value pair
func (l *PrettyLogger) KeyValue(key string, value interface{}) {
	fmt.Printf("  %s│%s %s%s:%s %v\n", BrightBlack, Reset, Cyan, key, Reset, value)
}

// SubDetail prints a sub-detail (double indented)
func (l *PrettyLogger) SubDetail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s│%s   %s%s%s\n", BrightBlack, Reset, Dim, msg, Reset)
}

// Highlight prints highlighted text
func (l *PrettyLogger) Highlight(text string) string {
	return fmt.Sprintf("%s%s%s", BrightYellow, text, Reset)
}

// URL prints a URL in a distinct color
func (l *PrettyLogger) URL(url string) string {
	return fmt.Sprintf("%s%s%s", BrightBlue, url, Reset)
}

// Number prints a number in a distinct color
func (l *PrettyLogger) Number(num interface{}) string {
	return fmt.Sprintf("%s%v%s", BrightMagenta, num, Reset)
}

// Percentage prints a percentage in a distinct color
func (l *PrettyLogger) Percentage(pct float64) string {
	return fmt.Sprintf("%s%.1f%%%s", BrightCyan, pct, Reset)
}

// FileSize formats file size with color
func (l *PrettyLogger) FileSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%s%d B%s", BrightMagenta, bytes, Reset)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%s%.2f KB%s", BrightMagenta, float64(bytes)/1024, Reset)
	} else {
		return fmt.Sprintf("%s%.2f MB%s", BrightMagenta, float64(bytes)/1024/1024, Reset)
	}
}

// Color prints a color hex with its RGB values
func (l *PrettyLogger) Color(hex string, r, g, b uint8, percentage float64) string {
	return fmt.Sprintf("%s%s%s (R:%s%3d%s, G:%s%3d%s, B:%s%3d%s) - %s",
		Bold, hex, Reset,
		BrightRed, r, Reset,
		BrightGreen, g, Reset,
		BrightBlue, b, Reset,
		l.Percentage(percentage))
}

// Divider prints a simple divider
func (l *PrettyLogger) Divider() {
	fmt.Printf("  %s│%s\n", BrightBlack, Reset)
}

// NewLine prints a new line
func (l *PrettyLogger) NewLine() {
	fmt.Println()
}
