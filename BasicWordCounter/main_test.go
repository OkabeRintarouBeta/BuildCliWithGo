package main

import (
	"bytes"
	"testing"
)

func TestWordCounter(t *testing.T) {
	b := bytes.NewBufferString("word1 word2 word3 word4\n")
	expected := 4
	result := count(b, 0)
	if result != expected {
		t.Errorf("Expected %d, got %d instead", expected, result)
	}
}

func TestLineCounter(t *testing.T) {
	b := bytes.NewBufferString("word1 word 2 \n word3 word4\n word5\n")
	expected := 3
	result := count(b, 1)
	if result != expected {
		t.Errorf("Expected %d, got %d instead", expected, result)
	}
}

func TestByteCounter(t *testing.T) {
	b := bytes.NewBufferString("word1 word 2 \n word3 word4\n word5\n")
	expected := 34
	result := count(b, 2)
	if result != expected {
		t.Errorf("Expected %d, got %d instead", expected, result)
	}
}
