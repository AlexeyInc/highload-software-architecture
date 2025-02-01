package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const elasticBulkURL = "http://localhost:9200/words_index/_bulk"
const elasticSearchURL = "http://localhost:9200/words_index/_search"

type Word struct {
	Word string `json:"word"`
}

func main() {
	file, err := os.Open("words.txt")
	if err != nil {
		log.Fatalf("Failed to open words file: %v", err)
	}
	defer file.Close()

	var bulkBuffer bytes.Buffer
	scanner := bufio.NewScanner(file)

	count := 0
	for scanner.Scan() {
		word := scanner.Text()
		meta := fmt.Sprintf(`{"index":{"_index":"words_index"}}`)
		doc, _ := json.Marshal(Word{Word: word})

		bulkBuffer.WriteString(meta + "\n")
		bulkBuffer.WriteString(string(doc) + "\n")

		count++
		if count%5000 == 0 { // Send in batches
			sendBulkRequest(bulkBuffer.String())
			bulkBuffer.Reset()
		}
	}

	if bulkBuffer.Len() > 0 {
		sendBulkRequest(bulkBuffer.String())
	}
	fmt.Println("Indexing completed.")

	//-----------------------------------

	http.HandleFunc("/search", searchHandler)
	log.Println("Server running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendBulkRequest(data string) {
	resp, err := http.Post(elasticBulkURL, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		log.Fatalf("Failed to send bulk request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Bulk insert status: %s\n", resp.Status)
}

type ESQuery struct {
	Query struct {
		Bool struct {
			Should []map[string]interface{} `json:"should"`
		} `json:"bool"`
	} `json:"query"`
}

type ESResponse struct {
	Hits struct {
		Hits []struct {
			Source struct {
				Word string `json:"word"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing 'q' parameter", http.StatusBadRequest)
		return
	}

	// Construct Query
	reqBody := ESQuery{}
	reqBody.Query.Bool.Should = []map[string]interface{}{
		{"match_phrase_prefix": map[string]interface{}{"word": query}},
		{"fuzzy": map[string]interface{}{"word": map[string]interface{}{"value": query, "fuzziness": "AUTO"}}},
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(elasticSearchURL, "application/json", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query ES: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var esResp ESResponse
	if err := json.NewDecoder(resp.Body).Decode(&esResp); err != nil {
		http.Error(w, "Failed to parse ES response", http.StatusInternalServerError)
		return
	}

	// Extract words
	var results []string
	for _, hit := range esResp.Hits.Hits {
		results = append(results, hit.Source.Word)
	}

	json.NewEncoder(w).Encode(results)
}
