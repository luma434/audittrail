package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/luma434/audittrail/internal/chain"
)

func newMux(svc *chain.Service, apiKey string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /events", requireAPIKey(apiKey, handleAppend(svc)))
	mux.HandleFunc("GET /events", requireAPIKey(apiKey, handleList(svc)))
	mux.HandleFunc("GET /verify", requireAPIKey(apiKey, handleVerify(svc)))
	return mux
}

func requireAPIKey(apiKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

type appendRequest struct {
	Payload json.RawMessage `json:"payload"`
}

func handleAppend(svc *chain.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req appendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if len(req.Payload) == 0 {
			http.Error(w, "payload is required", http.StatusBadRequest)
			return
		}

		entry, err := svc.AppendEvent(r.Context(), req.Payload, time.Now())
		if err != nil {
			http.Error(w, "failed to append event", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(entry)
	}
}

func handleList(svc *chain.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		if v := r.URL.Query().Get("limit"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = n
		}
		offset := 0
		if v := r.URL.Query().Get("offset"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				http.Error(w, "invalid offset", http.StatusBadRequest)
				return
			}
			offset = n
		}

		entries, err := svc.ListEvents(r.Context(), limit, offset)
		if err != nil {
			http.Error(w, "failed to list events", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}
}

type verifyResponse struct {
	Valid    bool   `json:"valid"`
	BrokenAt int    `json:"broken_at,omitempty"`
	Error    string `json:"error,omitempty"`
}

func handleVerify(svc *chain.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		brokenAt, err := svc.VerifyChain(r.Context())
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			json.NewEncoder(w).Encode(verifyResponse{Valid: false, BrokenAt: brokenAt, Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(verifyResponse{Valid: true})
	}
}
