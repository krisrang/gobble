package main

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/krisrang/gobble/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
	"github.com/krisrang/gobble/Godeps/_workspace/src/github.com/shiblon/entrogo/scrabble/index"
)

var (
	wwfFile       = "wordswithfriends.mealy"
	scrabbleFile  = "TWL06.mealy"
	wwfIndex      index.Index
	scrabbleIndex index.Index

	scrabbleScores = map[rune]int{
		'A': 1, 'B': 3, 'C': 3, 'D': 2, 'E': 1, 'F': 4, 'G': 2, 'H': 4, 'I': 1,
		'J': 8, 'K': 5, 'L': 1, 'M': 3, 'N': 1, 'O': 1, 'P': 3, 'Q': 10, 'R': 1,
		'S': 1, 'T': 1, 'U': 1, 'V': 4, 'W': 4, 'X': 8, 'Y': 4, 'Z': 10,
	}

	wwfScores = map[rune]int{
		'A': 1, 'B': 4, 'C': 4, 'D': 2, 'E': 1, 'F': 4, 'G': 3, 'H': 3, 'I': 1,
		'J': 10, 'K': 5, 'L': 2, 'M': 4, 'N': 2, 'O': 1, 'P': 4, 'Q': 10, 'R': 1,
		'S': 1, 'T': 1, 'U': 2, 'V': 5, 'W': 4, 'X': 8, 'Y': 3, 'Z': 10,
	}
)

type Word struct {
	String string `json:"string,omitempty"`
	Score  int    `json:"score,omitempty"`
}

func NewWord(word string, wwf bool) Word {
	result := Word{
		String: word,
		Score:  0,
	}

	for _, v := range word {
		if wwf {
			result.Score += wwfScores[rune(v)]
		} else {
			result.Score += scrabbleScores[rune(v)]
		}
	}

	return result
}

type ByLength []Word

func (s ByLength) Len() int {
	return len(s)
}
func (s ByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByLength) Less(i, j int) bool {
	return len(s[i].String) > len(s[j].String)
}

func TileCounts(available string) (counts map[byte]int) {
	counts = make(map[byte]int)
	for _, c := range strings.ToUpper(available) {
		counts[byte(c)]++
	}
	return
}

func PossibleWords(idx index.Index, available map[byte]int, wwf bool) []Word {
	allowed := index.NewUnanchoredAllowedInfo(
		[]string{".", ".", ".", ".", ".", ".", "."},
		[]bool{true, true, true, true, true, true, true},
		available)

	out := make([]Word, 0)
	for seq := range idx.ConstrainedSequences(allowed) {
		out = append(out, NewWord(string(seq), wwf))
	}

	sort.Sort(ByLength(out))

	return out
}

func GetScrabble(w rest.ResponseWriter, r *rest.Request) {
	tiles := TileCounts(strings.Replace(r.PathParam("tiles"), "+", ".", -1))
	words := PossibleWords(scrabbleIndex, tiles, true)

	result := make([]Word, 0)
	for _, v := range words {
		result = append(result, v)
	}
	w.WriteJson(&result)
}

func GetWords(w rest.ResponseWriter, r *rest.Request) {
	tiles := TileCounts(strings.Replace(r.PathParam("tiles"), "+", ".", -1))
	words := PossibleWords(wwfIndex, tiles, false)

	result := make([]Word, 0)
	for _, v := range words {
		result = append(result, v)
	}
	w.WriteJson(&result)
}

type hijack404 struct {
	http.ResponseWriter
	R         *http.Request
	Handle404 func(w http.ResponseWriter, r *http.Request) bool
}

func (h *hijack404) WriteHeader(code int) {
	if 404 == code && h.Handle404(h.ResponseWriter, h.R) {
		panic(h)
	}
	h.ResponseWriter.WriteHeader(code)
}

// Handle404 will pass any 404's from the handler to the handle404
// function. If handle404 returns true, the response is considered complete,
// and the processing by handler is aborted.
func Handle404(handler http.Handler, handle404 func(w http.ResponseWriter, r *http.Request) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijack := &hijack404{ResponseWriter: w, R: r, Handle404: handle404}
		defer func() {
			if p := recover(); p != nil {
				if p == hijack {
					return
				}
				panic(p)
			}
		}()
		handler.ServeHTTP(hijack, r)
	})
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
		rest.Get("/scrabble/:tiles", GetScrabble),
		rest.Get("/wwf/:tiles", GetWords),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))
	http.Handle("/", Handle404(http.FileServer(http.Dir("./web/build")), func(w http.ResponseWriter, r *http.Request) bool {
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			w.Header().Set("Content-Type", "text/html")
			http.ServeFile(w, r, "./web/build/index.html")
			return true
		}

		return false
	}))

	log.Println("Listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
