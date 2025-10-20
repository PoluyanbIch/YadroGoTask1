package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const dir = "./upload"

func main() {
	http.HandleFunc("/files", filesHandler)
	http.HandleFunc("/files/", filenameHandler)

	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		filesHandlerPost(w, r)
	case http.MethodGet:
		filesHandlerGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
		http.Error(w, fmt.Sprintf("Error in making dir: %s", err.Error()), http.StatusInternalServerError)
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

func filesHandlerGet(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading Directory Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	for _, file := range files {
		w.Write([]byte(file.Name() + "\n"))
	}
	w.WriteHeader(http.StatusOK)
}

func filenameHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	filename := strings.TrimPrefix(path, "/files/")

	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		filenameHandlerPut(w, r, filename)
	case http.MethodGet:
		return
	case http.MethodDelete:
		return
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func filenameHandlerPut(w http.ResponseWriter, r *http.Request, filename string) {
	filePath := filepath.Join(dir, filename)

	if _, err := os.Stat(filePath); err != nil {
		http.Error(w, fmt.Sprintf("File not found Error: %s", err.Error()), http.StatusNotFound)
		return
	}

	prevFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		http.Error(w, fmt.Sprintf("Open file Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer prevFile.Close()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, fmt.Sprintf("ParseMultipleForm Error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("FormFile Error: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if _, err := io.Copy(prevFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Copying Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
