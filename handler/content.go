package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"content-public-api/store"
)

const (
	cacheTTL     = 15 * time.Second
	queryTimeout = 5 * time.Second
)

type ContentHandler struct {
	store *store.ContentStore
	cache *gocache.Cache
}

func NewContentHandler(s *store.ContentStore) *ContentHandler {
	return &ContentHandler{
		store: s,
		cache: gocache.New(cacheTTL, 2*cacheTTL),
	}
}

func (h *ContentHandler) GetSections(w http.ResponseWriter, r *http.Request) {
	const cacheKey = "sections"

	if cached, found := h.cache.Get(cacheKey); found {
		writeJSON(w, cached)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	result, err := h.store.GetGroupedSections(ctx)
	if err != nil {
		log.Printf("GetSections error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.cache.Set(cacheKey, result, gocache.DefaultExpiration)
	writeJSON(w, result)
}

func (h *ContentHandler) SearchContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q is required", http.StatusBadRequest)
		return
	}

	cacheKey := "search:q=" + q

	if cached, found := h.cache.Get(cacheKey); found {
		writeJSON(w, cached)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	rows, err := h.store.SearchContent(ctx, q)
	if err != nil {
		log.Printf("SearchContent error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.cache.Set(cacheKey, rows, gocache.DefaultExpiration)
	writeJSON(w, rows)
}

func (h *ContentHandler) GetContentBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "slug is required", http.StatusBadRequest)
		return
	}

	cacheKey := "slug:" + slug

	if cached, found := h.cache.Get(cacheKey); found {
		writeJSON(w, cached)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	content, err := h.store.GetContentBySlug(ctx, slug)
	if err != nil {
		log.Printf("GetContentBySlug error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if content == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.cache.Set(cacheKey, content, gocache.DefaultExpiration)
	writeJSON(w, content)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
