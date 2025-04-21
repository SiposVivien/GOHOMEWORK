package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// for atomic
var timestamp int64

func handleTimestamp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	switch r.Method {
	case http.MethodPost:
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid timestamp", http.StatusBadRequest)
			return
		}
		atomic.StoreInt64(&timestamp, ts)
	case http.MethodGet:
		ts := atomic.LoadInt64(&timestamp)
		_, _ = fmt.Fprintf(w, "%d", ts)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/timestamp", handleTimestamp)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	//waiting for the server
	time.Sleep(100 * time.Millisecond)

	//the client is send a timestamp
	now := time.Now().Unix()
	_, err := http.Post("http://localhost:8080/timestamp", "text/plain", io.NopCloser(strings.NewReader(strconv.FormatInt(now, 10))))
	if err != nil {
		log.Fatal("POST error:", err)
	}

	//Kliens timestamp lekérés
	resp, err := http.Get("http://localhost:8080/timestamp")
	if err != nil {
		log.Fatal("POST error:", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println(string(body)) //csak ez fog megjelenni couton
}
