package utils

import (
	"bufio"
	"os"
)

func ParseSymbols() ([]string, error) {
	file, err := os.Open("./symbols.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	symbols := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		symbols = append(symbols, scanner.Text())
	}
	return symbols, nil
}
