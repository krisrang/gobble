package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shiblon/entrogo/scrabble/index"
)

var (
	recognizerFile = "wordswithfriends.mealy"
)

func TileCounts(available string) (counts map[byte]int) {
	counts = make(map[byte]int)
	for _, c := range strings.ToUpper(available) {
		counts[byte(c)]++
	}
	return
}

func PossibleWords(idx index.Index, available map[byte]int) <-chan string {
	allowed := index.NewUnanchoredAllowedInfo(
		[]string{".", ".", ".", ".", ".", ".", "."},
		[]bool{true, true, true, true, true, true, true},
		available)

	out := make(chan string)

	go func() {
		defer close(out)
		for seq := range idx.ConstrainedSequences(allowed) {
			out <- string(seq)
		}
	}()

	return out
}

func main() {
	rfile, err := os.Open(recognizerFile)
	if err != nil {
		log.Fatalf("Failed to open '%v': %v", recognizerFile, err)
	}
	idx, err := index.ReadFrom(rfile)
	if err != nil {
		log.Fatalf("Failed to read recognizer from '%v': %v", recognizerFile, err)
	}

	available := TileCounts("abcdef")
	allwords := make([]string, 0, 500)
	for word := range PossibleWords(idx, available) {
		allwords = append(allwords, word)
	}

	for _, w := range allwords {
		fmt.Println(w)
	}
}
