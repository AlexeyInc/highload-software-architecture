package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

var cache sync.Map

func getImage(w http.ResponseWriter, r *http.Request) {
	imageName := r.URL.Path[len("/image/"):]

	if data, found := cache.Load(imageName); found {
		log.Println("Cache hit for", imageName)
		w.Write(data.([]byte))
		return
	}

	filePath := fmt.Sprintf("./images/%s", imageName)
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	cache.Store(imageName, data)
	log.Println("Cached", imageName)
	w.Write(data)
}

func main() {
	http.HandleFunc("/image/", getImage)
	log.Println("Image server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
