package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

type Config struct {
	countMode int
}

func main() {
	config := Config{}
	flag.IntVar(&config.countMode, "mode", 0, "Counting Mode:\n\t 0--count words\t 1--count lines\t 2--count bytes")
	flag.Parse()
	fmt.Println(count(os.Stdin, config.countMode))
}

func count(r io.Reader, countMode int) int {

	scanner := bufio.NewScanner(r)
	switch countMode {
	case 0:
		scanner.Split(bufio.ScanWords)
	case 2:
		scanner.Split(bufio.ScanBytes)
	}

	ct := 0
	for scanner.Scan() {
		ct++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return ct
}
