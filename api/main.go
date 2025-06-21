package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
)

type Input struct {
	Text string `json:"text"`
}

func main() {
	ichiran.MustInit()
	defer ichiran.Close()

	http.HandleFunc("/analyze", analyzeHandler)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST", http.StatusMethodNotAllowed)
		return
	}

	var in Input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	tokens, err := ichiran.Analyze(in.Text)
	if err != nil {
		http.Error(w, "Analyze failed", http.StatusInternalServerError)
		return
	}

	out := map[string]interface{}{
		"tokenized":      tokens.Tokenized(),
		"tokenizedParts": tokens.TokenizedParts(),
		"kana":           tokens.Kana(),
		"roman":          tokens.Roman(),
		"glossParts":     tokens.GlossParts(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
