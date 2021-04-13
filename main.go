package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Searcher struct {
	records []Record
}

type Record struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	ThumbURL  string   `json:"thumb_url"`
	Tags      []string `json:"tags"`
	UpdatedAt int64    `json:"updated_at"`
	ImageURLs []string `json:"image_urls"`
}

func main() {
	// initialize searcher
	searcher := &Searcher{}
	err := searcher.Load("data.gz")
	if err != nil {
		log.Fatalf("unable to load search data due: %v", err)
	}
	// define http handlers
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/search", handleSearch(searcher))
	// start server
	port := os.Getenv("PORT")
	if err != nil {
		port = "3000"
	}
	fmt.Printf("Server is listening on %v...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatalf("unable to start server due: %v", err)
	}
}

func handleSearch(s *Searcher) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// fetch query string from query params
			q := r.URL.Query().Get("q")
			if len(q) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Well, You can search another thing here..."))
				return
			}
			// search relevant records
			records, err := s.Search(q)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Oops, Well this is Unexpected..."))
				return
			}
			// output success response
			buf := new(bytes.Buffer)
			encoder := json.NewEncoder(buf)
			encoder.Encode(records)
			w.Header().Set("Content-Type", "application/json")
			w.Write(buf.Bytes())
		},
	)
}

func (s *Searcher) Load(filepath string) error {
	// open file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("unable to open source file due: %v", err)
	}
	defer file.Close()
	// read as gzip
	reader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("unable to initialize gzip reader due: %v", err)
	}
	// read the reader using scanner to contstruct records
	var records []Record
	cs := bufio.NewScanner(reader)
	for cs.Scan() {
		var r Record
		err = json.Unmarshal(cs.Bytes(), &r)
		if err != nil {
			continue
		}
		records = append(records, r)
	}
	s.records = records

	return nil
}

func (s *Searcher) Search(query string) ([]Record, error) {
	var result []Record
	for _, record := range s.records {
		if strings.Contains(record.Title, query) || strings.Contains(record.Content, query) {
			result = append(result, record)
		}
	}
	return result, nil
}
