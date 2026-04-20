package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// SessionHeader carries the API session ID on every authenticated call.
const SessionHeader = "X-Keycraft-Session"

// NewRouter builds the HTTP mux for the Service. Wrap the result in
// RequireBearer (or equivalent auth middleware) before serving externally.
func NewRouter(svc *Service) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		dto, err := svc.OpenSession()
		writeJSON(w, dto, err)
	})

	mux.HandleFunc("DELETE /v1/sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := svc.CloseSession(id); err != nil {
			writeErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /v1/corpora", func(w http.ResponseWriter, r *http.Request) {
		names, err := svc.ListCorpora()
		writeJSON(w, map[string]any{"corpora": names}, err)
	})

	mux.HandleFunc("POST /v1/corpora/load", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := sessionID(w, r)
		if !ok {
			return
		}
		var ref CorpusRef
		if err := decodeJSON(r, &ref); err != nil {
			writeErr(w, err)
			return
		}
		dto, err := svc.DescribeCorpus(sid, ref)
		writeJSON(w, dto, err)
	})

	mux.HandleFunc("GET /v1/layouts", func(w http.ResponseWriter, r *http.Request) {
		names, err := svc.ListLayouts()
		writeJSON(w, map[string]any{"layouts": names}, err)
	})

	mux.HandleFunc("GET /v1/layouts/{name}", func(w http.ResponseWriter, r *http.Request) {
		dto, err := svc.GetLayout(r.PathValue("name"))
		writeJSON(w, dto, err)
	})

	mux.HandleFunc("POST /v1/layouts/{name}/flip", func(w http.ResponseWriter, r *http.Request) {
		dto, err := svc.FlipLayout(r.PathValue("name"))
		writeJSON(w, dto, err)
	})

	mux.HandleFunc("POST /v1/analyse", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := sessionID(w, r)
		if !ok {
			return
		}
		var req AnalyseRequest
		if err := decodeJSON(r, &req); err != nil {
			writeErr(w, err)
			return
		}
		dto, err := svc.Analyse(sid, req)
		writeJSON(w, dto, err)
	})

	mux.HandleFunc("POST /v1/view", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := sessionID(w, r)
		if !ok {
			return
		}
		var req ViewRequest
		if err := decodeJSON(r, &req); err != nil {
			writeErr(w, err)
			return
		}
		dto, err := svc.View(sid, req)
		writeJSON(w, dto, err)
	})

	mux.HandleFunc("POST /v1/rank", func(w http.ResponseWriter, r *http.Request) {
		sid, ok := sessionID(w, r)
		if !ok {
			return
		}
		var req RankRequest
		if err := decodeJSON(r, &req); err != nil {
			writeErr(w, err)
			return
		}
		dto, err := svc.Rank(sid, req)
		writeJSON(w, dto, err)
	})

	return mux
}

func sessionID(w http.ResponseWriter, r *http.Request) (string, bool) {
	sid := r.Header.Get(SessionHeader)
	if sid == "" {
		http.Error(w, "missing "+SessionHeader+" header", http.StatusBadRequest)
		return "", false
	}
	return sid, true
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, body any, err error) {
	if err != nil {
		writeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrSessionNotFound),
		errors.Is(err, ErrSessionExpired):
		status = http.StatusNotFound
	case errors.Is(err, ErrSessionLimit),
		errors.Is(err, ErrBusy):
		status = http.StatusTooManyRequests
	case errors.Is(err, ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	msg := err.Error()
	if strings.ContainsRune(msg, '\n') {
		msg = strings.ReplaceAll(msg, "\n", " ")
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
