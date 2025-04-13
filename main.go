package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// In-memory storage for receipt points.
var receiptPointsStore = make(map[string]int64)
var storeMutex = &sync.RWMutex{}

// API error messages.
const badRequestMsg = "The receipt is invalid."
const notFoundMsg = "No receipt found for that ID."

// Handles POST /receipts/process requests.
func processReceiptHandler(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {

	var receipt Receipt
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&receipt); err != nil {
		logger.Warn("Failed to decode receipt JSON", slog.Any("error", err))
		errorResponse(w, http.StatusBadRequest, badRequestMsg, logger)
		return
	}

	validatedData, err := validateAndParseReceipt(&receipt)
	if err != nil {
		logger.Warn("Receipt validation failed", slog.Any("error", err), slog.String("retailer", receipt.Retailer))
		errorResponse(w, http.StatusBadRequest, badRequestMsg, logger)
		return
	}

	points := calculatePoints(validatedData)
	id := uuid.NewString()

	storeMutex.Lock()
	receiptPointsStore[id] = points
	storeMutex.Unlock()

	logger.Info("Receipt processed", slog.String("id", id), slog.Int64("points", points), slog.String("retailer", validatedData.Retailer))

	type ProcessResponse struct {
		ID string `json:"id"`
	}
	jsonResponse(w, http.StatusOK, ProcessResponse{ID: id}, logger)
}

// Handles GET /receipts/{id}/points requests.
func getPointsHandler(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	id := r.PathValue("id")

	if id == "" || !idPatternRegex.MatchString(id) {
		logger.Warn("Invalid ID format requested", slog.String("requested_id", id))
		errorResponse(w, http.StatusNotFound, notFoundMsg, logger)
		return
	}

	storeMutex.RLock()
	points, found := receiptPointsStore[id]
	storeMutex.RUnlock()

	if !found {
		logger.Warn("Receipt ID not found", slog.String("id", id))
		errorResponse(w, http.StatusNotFound, notFoundMsg, logger)
		return
	}

	logger.Info("Points retrieved", slog.String("id", id), slog.Int64("points", points))

	type PointsResponse struct {
		Points int64 `json:"points"`
	}
	jsonResponse(w, http.StatusOK, PointsResponse{Points: points}, logger)
}

// main is the application entry point.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	mux := http.NewServeMux()

	// Register endpoint handlers
	mux.HandleFunc("POST /receipts/process", func(w http.ResponseWriter, r *http.Request) {
		processReceiptHandler(w, r, logger)
	})
	mux.HandleFunc("GET /receipts/{id}/points", func(w http.ResponseWriter, r *http.Request) {
		getPointsHandler(w, r, logger)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Receipt Processor API Ready"))
	})

	// Determine port or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure and start server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Server starting...", slog.String("port", port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
