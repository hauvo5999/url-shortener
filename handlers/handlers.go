package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	
	"url-shortener/models"
	"url-shortener/storage"
)

type URLHandler struct {
	storage *storage.MemoryStorage
}

func NewURLHandler(storage *storage.MemoryStorage) *URLHandler {
	return &URLHandler{storage: storage}
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Validate URL
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	
	// Generate short code
	shortCode, err := h.generateShortCode()
	if err != nil {
		http.Error(w, "Failed to generate short code", http.StatusInternalServerError)
		return
	}
	
	// Create URL model
	urlModel := &models.URL{
		ID:          uuid.New().String(),
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}
	
	// Store in database
	if err := h.storage.Store(urlModel); err != nil {
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}
	
	// Return response
	response := models.ShortenResponse{
		ShortCode:   shortCode,
		ShortURL:    fmt.Sprintf("http://localhost:8080/%s", shortCode),
		OriginalURL: req.URL,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]
	
	// Get URL from storage
	urlModel, err := h.storage.GetByShortCode(shortCode)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}
	
	// Increment access count
	if err := h.storage.IncrementAccess(shortCode); err != nil {
		log.Printf("Failed to increment access count: %v", err)
	}
	
	// Redirect
	http.Redirect(w, r, urlModel.OriginalURL, http.StatusMovedPermanently)
}

func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]
	
	urlModel, err := h.storage.GetByShortCode(shortCode)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}
	
	response := models.StatsResponse{
		ShortCode:   urlModel.ShortCode,
		OriginalURL: urlModel.OriginalURL,
		CreatedAt:   urlModel.CreatedAt,
		AccessCount: urlModel.AccessCount,
		LastAccess:  urlModel.LastAccess,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *URLHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	stats := h.storage.GetStats()
	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now(),
		"stats":  stats,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *URLHandler) generateShortCode() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	
	return string(result), nil
}