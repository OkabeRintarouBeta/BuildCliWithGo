package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Println(count(os.Stdin))
}

func count(r io.Reader) int {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	ct := 0
	for scanner.Scan() {
		ct++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return ct
}
