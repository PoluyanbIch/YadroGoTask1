package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port int `yaml:"port" env:"HELLO_PORT" env-default:"8080"`
}

var cfg Config

func loadConfig() error {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configPath != "" {
		return cleanenv.ReadConfig(*configPath, &cfg)
	}
	return cleanenv.ReadEnv(&cfg)
}

const dir = "./upload"

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)

	http.HandleFunc("POST /files", filesHandlerPost)
	http.HandleFunc("GET /files", filesHandlerGet)
	http.HandleFunc("PUT /files/{filename}", filenameHandlerPut)
	http.HandleFunc("GET /files/{filename}", filenameHandlerGet)
	http.HandleFunc("DELETE /files/{filename}", filenameHandlerDelete)

	fmt.Printf("Server is listening on port %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func filesHandlerPost(w http.ResponseWriter, r *http.Request) {
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

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in making dir: %v", err), http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(dir, header.Filename)

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

func filesHandlerGet(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(dir)
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

func filenameHandlerPut(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(dir, filename)

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

func filenameHandlerGet(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(dir, filename)

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

func filenameHandlerDelete(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(dir, filename)

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
