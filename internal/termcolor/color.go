package termcolor

// ANSI color escape codes
const (
	Reset  = "\u001b[0m"
	Red    = "\u001b[31m"
	Green  = "\u001b[32m"
	Yellow = "\u001b[33m"
	Blue   = "\u001b[34m"
	Purple = "\u001b[35m"
	Cyan   = "\u001b[36m"
	White  = "\u001b[37m"
	Bold   = "\u001b[1m"
)

// Text wraps text in the specified color and resets afterward
func Text(text string, color string) string {
	return color + text + Reset
}