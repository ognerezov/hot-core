package console

import "fmt"

var (
	// Reset is the ANSI escape code to reset formatting.
	Reset = "\033[0m"
	// Red is the ANSI escape code for red text.
	Red = "\033[31m"
	// Green is the ANSI escape code for green text.
	Green = "\033[32m"
	// Yellow is the ANSI escape code for yellow text.
	Yellow = "\033[33m"
	// Blue is the ANSI escape code for blue text.
	Blue = "\033[34m"
	// Magenta is the ANSI escape code for magenta text.
	Magenta = "\033[35m"
	// Cyan is the ANSI escape code for cyan text.
	Cyan = "\033[36m"
	// Gray is the ANSI escape code for gray text.
	Gray = "\033[37m"
	// White is the ANSI escape code for white text.
	White = "\033[97m"
)

// ColorPrintln prints a string to standard output with the specified ANSI color.
func ColorPrintln(str string, color string) {
	fmt.Println(color + str + Reset)
}

// BluePrintln prints a string to standard output in blue.
func BluePrintln(str string) {
	ColorPrintln(str, Blue)
}

// MagentaPrintln prints a string to standard output in magenta.
func MagentaPrintln(str string) {
	ColorPrintln(str, Magenta)
}

// CyanPrintln prints a string to standard output in cyan.
func CyanPrintln(str string) {
	ColorPrintln(str, Cyan)
}

// GreenPrintln prints a string to standard output in green.
func GreenPrintln(str string) {
	ColorPrintln(str, Green)
}

// RedPrintln prints a string to standard output in red.
func RedPrintln(str string) {
	ColorPrintln(str, Red)
}

// YellowPrintln prints a string to standard output in yellow.
func YellowPrintln(str string) {
	ColorPrintln(str, Yellow)
}

// RedPrint prints a string to standard output in red (alias for RedPrintln).
func RedPrint(str string) {
	ColorPrintln(str, Red)
}
