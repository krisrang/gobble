package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/krisrang/gobble/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	"github.com/krisrang/gobble/Godeps/_workspace/src/github.com/shiblon/entrogo/scrabble/index"
)

var (
	wwfFile       = "wordswithfriends.mealy"
	scrabbleFile  = "TWL06.mealy"
	wwfIndex      index.Index
	scrabbleIndex index.Index
)

func tileCounts(available string) (counts map[byte]int) {
	counts = make(map[byte]int)
	for _, c := range strings.ToUpper(available) {
		counts[byte(c)]++
	}
	return
}

func possibleWords(idx index.Index, available map[byte]int) <-chan string {
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

func Index(w rest.ResponseWriter, r *rest.Request) {
	rest.NotFound(w, r)
}

func GetScrabble(w rest.ResponseWriter, r *rest.Request) {
	tiles := tileCounts(strings.Replace(r.PathParam("tiles"), "+", ".", -1))
	words := possibleWords(scrabbleIndex, tiles)

	result := make([]string, 0)
	for v := range words {
		result = append(result, v)
	}
	w.WriteJson(&result)
}

func GetWords(w rest.ResponseWriter, r *rest.Request) {
	tiles := tileCounts(strings.Replace(r.PathParam("tiles"), "+", ".", -1))
	words := possibleWords(wwfIndex, tiles)

	result := make([]string, 0)
	for v := range words {
		result = append(result, v)
	}
	w.WriteJson(&result)
}

func main() {
	wwffile, err := os.Open(wwfFile)
	if err != nil {
		log.Fatalf("Failed to open '%v': %v", wwfFile, err)
	}

	wwfIndex, err = index.ReadFrom(wwffile)
	if err != nil {
		log.Fatalf("Failed to read recognizer from '%v': %v", wwfFile, err)
	}

	scrabblefile, err := os.Open(scrabbleFile)
	if err != nil {
		log.Fatalf("Failed to open '%v': %v", scrabbleFile, err)
	}

	scrabbleIndex, err = index.ReadFrom(scrabblefile)
	if err != nil {
		log.Fatalf("Failed to read recognizer from '%v': %v", scrabbleFile, err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/", Index),
		rest.Get("/scrabble/:tiles", GetScrabble),
		rest.Get("/wwf/:tiles", GetWords),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	log.Println("Listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, api.MakeHandler()))
}
