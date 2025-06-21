package handler

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tassa-yoniso-manasi-karoto/go-ichiran"
	"net/http"
	"time"
)

type TextHandler struct {
	logger *logrus.Logger
}

type AnalyzeRequest struct {
	Text string `json:"text"`
}

type AnalyzeResponse struct {
	Tokenized              string   `json:"tokenized"`
	TokenizedParts         []string `json:"tokenized_parts"`
	Kana                   string   `json:"kana"`
	KanaParts              []string `json:"kana_parts"`
	Roman                  string   `json:"roman"`
	RomanParts             []string `json:"roman_parts"`
	SelectiveTranslit      string   `json:"selective_translit"`
	SelectiveTranslitParts string   `json:"selective_translit_parts"`
	GlossParts             []string `json:"gloss_parts"`
}

func NewTextHandler(logger *logrus.Logger) *TextHandler {
	return &TextHandler{logger: logger}
}

func (h *TextHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		h.logger.Error("Text field is required")
		http.Error(w, "Text field is required", http.StatusBadRequest)
		return
	}

	tokens, err := ichiran.AnalyzeWithContext(ctx, req.Text)
	if err != nil {
		h.logger.Errorf("Analysis failed: %v", err)
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	// Perform selective transliteration (top 1000 kanji)
	tlit, err := tokens.SelectiveTranslit(1000)
	if err != nil {
		h.logger.Errorf("Selective transliteration failed: %v", err)
		http.Error(w, "Selective transliteration failed", http.StatusInternalServerError)
		return
	}

	tlitTokenized, err := tokens.SelectiveTranslitTokenized(1000)
	if err != nil {
		h.logger.Errorf("Selective tokenized transliteration failed: %v", err)
		http.Error(w, "Selective tokenized transliteration failed", http.StatusInternalServerError)
		return
	}

	resp := AnalyzeResponse{
		Tokenized:              tokens.Tokenized(),
		TokenizedParts:         tokens.TokenizedParts(),
		Kana:                   tokens.Kana(),
		KanaParts:              tokens.KanaParts(),
		Roman:                  tokens.Roman(),
		RomanParts:             tokens.RomanParts(),
		SelectiveTranslit:      tlit,
		SelectiveTranslitParts: tlitTokenized,
		GlossParts:             tokens.GlossParts(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
