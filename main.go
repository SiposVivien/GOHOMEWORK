package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

//azért var típusú, hogy mindenki elérje, és hogy a lifetime-ja a program végéig lefusson
var (
	timestampChan = make(chan time.Time, 1) // Pufferelt csatorna
)
//responsewriter: interface ami lehetővé teszi a http válasz írását
func saveTimestamp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	unixTime, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil || unixTime < 0 { // Negatív érték ellenőrzése
		http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
		return
	}

	timestamp := time.Unix(unixTime, 0)
	select {
	case timestampChan <- timestamp:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Timestamp saved"))
	default:
		http.Error(w, "Storage busy", http.StatusServiceUnavailable)
	}
}

func getTimestamp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	select {
	case timestamp := <-timestampChan:
		// Visszarakjuk és küldjük az értéket
		timestampChan <- timestamp
		w.Write([]byte(strconv.FormatInt(timestamp.Unix(), 10)))
	default:
		http.Error(w, "No timestamp available", http.StatusNotFound)
	}
}

func main() {
	// Kezdeti érték beállítása
	timestampChan <- time.Now()

	http.HandleFunc("/timestamp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		switch r.Method {
		case http.MethodPost:
			saveTimestamp(w, r)
		case http.MethodGet:
			getTimestamp(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Szerver indítása
	go func() {
		fmt.Println("Server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Kliens teszt
	client := &http.Client{}
	currentTime := time.Now().Unix()

	// Mentés
	resp, err := client.Post("http://localhost:8080/timestamp", "text/plain",
		io.NopCloser(bytes.NewReader([]byte(strconv.FormatInt(currentTime, 10)))))
	if err != nil {
		fmt.Println("Error saving timestamp:", err)
		os.Exit(1)
	}
	resp.Body.Close()

	// Olvasás
	resp, err = client.Get("http://localhost:8080/timestamp")
	if err != nil {
		fmt.Println("Error getting timestamp:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	fmt.Println(string(body))
}
