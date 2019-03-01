package util

import (
	"bufio"
	"os"
)

// PipeGiven checks if command is piped
func PipeGiven() bool {
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// LinesInFile return all lines from the given file
func LinesInFile(fileName string) []string {
	f, _ := os.Open(fileName)
	scanner := bufio.NewScanner(f)
	return readLines(scanner)
}

// LinesInStdin return all lines from stdin
func LinesInStdin() []string {
	scanner := bufio.NewScanner(os.Stdin)
	return readLines(scanner)
}

func readLines(scanner *bufio.Scanner) (result []string) {
	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}
	return
}
