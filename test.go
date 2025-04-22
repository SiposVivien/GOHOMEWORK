package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSaveAndGetTimestamp(t *testing.T) {
	// Inicializáljuk a csatornát
	timestampChan = make(chan time.Time, 1)
	timestampChan <- time.Time{}

	// Teszt szerver létrehozása
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		switch r.Method {
		case http.MethodPost:
			saveTimestamp(w, r)
		case http.MethodGet:
			getTimestamp(w, r)
		}
	}))
	defer ts.Close()

	// Mentés tesztelése
	currentTime := time.Now().Unix()
	resp, err := http.Post(ts.URL+"/timestamp", "text/plain",
		strings.NewReader(strconv.FormatInt(currentTime, 10)))
	if err != nil {
		t.Fatalf("Failed to save timestamp: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Olvasás tesztelése
	resp, err = http.Get(ts.URL + "/timestamp")
	if err != nil {
		t.Fatalf("Failed to get timestamp: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
}

func TestInvalidMethods(t *testing.T) {
	// Inicializáljuk a csatornát
	timestampChan = make(chan time.Time, 1)
	timestampChan <- time.Time{}

	// Teszt szerver létrehozása
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		switch r.Method {
		case http.MethodPost:
			saveTimestamp(w, r)
		case http.MethodGet:
			getTimestamp(w, r)
		}
	}))
	defer ts.Close()

	// Nem támogatott HTTP metódus tesztelése
	resp, err := http.Head(ts.URL + "/timestamp")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status MethodNotAllowed, got %v", resp.StatusCode)
	}
}
