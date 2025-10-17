package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const dir = "./upload"

func main() {
	http.HandleFunc("/files", filesHandler)

	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filesHandlerPost(w, r)
}

func filesHandlerPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, fmt.Sprintf("ParseMultipleForm Error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("FormFile Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in male dir: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(dir, header.Filename)

	if _, err := os.Stat(filePath); err == nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("File already exist"))
		return
	}

	newFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Creation Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Copying Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("File upploading success"))
}
