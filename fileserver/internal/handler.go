package internal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type FileHandler struct {
	dir string
}

func NewFileHandler(dir string) *FileHandler {
	return &FileHandler{dir: dir}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, fmt.Sprintf("ParseMultipleForm Error: %v", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("FormFile Error: %v", err), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	err = os.MkdirAll(h.dir, 0755)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in making dir: %v", err), http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(h.dir, header.Filename)

	if _, err := os.Stat(filePath); err == nil {
		w.WriteHeader(http.StatusConflict)
		if _, err := w.Write([]byte("File already exist")); err != nil {
			log.Printf("Write Error: %v", err)
		}
		return
	}

	newFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Creation Error: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := newFile.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	_, err = io.Copy(newFile, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Copying Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(header.Filename)); err != nil {
		log.Printf("Write Error: %v", err)
	}
}

func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(h.dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading Directory Error: %v", err), http.StatusInternalServerError)
		return
	}
	for _, file := range files {
		if _, err := w.Write([]byte(file.Name() + "\n")); err != nil {
			log.Printf("Write Error: %v", err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.dir, filename)

	if _, err := os.Stat(filePath); err != nil {
		http.Error(w, fmt.Sprintf("File not found Error: %v", err), http.StatusNotFound)
		return
	}

	prevFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		http.Error(w, fmt.Sprintf("Open file Error: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := prevFile.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, fmt.Sprintf("ParseMultipleForm Error: %v", err), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("FormFile Error: %v", err), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	if _, err := io.Copy(prevFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Copying Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *FileHandler) Update(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.dir, filename)

	if _, err := os.Stat(filePath); err != nil {
		http.Error(w, fmt.Sprintf("File not found Error: %v", err), http.StatusNotFound)
		return
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		http.Error(w, fmt.Sprintf("Open file Error: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, fmt.Sprintf("Copying file Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.dir, filename)

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, fmt.Sprintf("File stat Error: %v", err), http.StatusInternalServerError)
		return
	}

	if err := os.Remove(filePath); err != nil {
		http.Error(w, fmt.Sprintf("Delete Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
