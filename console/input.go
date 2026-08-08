package console

import (
	"bufio"
	"os"
	"strings"
)

// ReadStr prints a prompt and reads a string from standard input.
func ReadStr(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	BluePrintln(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSuffix(text, "\n")
}
