package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	http.HandleFunc("/files", filesHandler)
	http.HandleFunc("/files/", filenameHandler)

	fmt.Printf("Server is listening on port %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %s", err.Error())
	}
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
		if _, err := w.Write([]byte("File already exist")); err != nil {
			log.Printf("Write Error: %s", err.Error())
		}
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
	if _, err := w.Write([]byte("File upploading success")); err != nil {
		log.Printf("Write Error: %s", err.Error())
	}
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
		filenameHandlerGet(w, r, filename)
	case http.MethodDelete:
		filenameHandlerDelete(w, r, filename)
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

func filenameHandlerGet(w http.ResponseWriter, r *http.Request, filename string) {
	filePath := filepath.Join(dir, filename)

	if _, err := os.Stat(filePath); err != nil {
		http.Error(w, fmt.Sprintf("File not found Error: %s", err.Error()), http.StatusNotFound)
		return
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read file Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(fileContent); err != nil {
		log.Printf("Write Error: %s", err.Error())
	}
}

func filenameHandlerDelete(w http.ResponseWriter, r *http.Request, filename string) {
	filePath := filepath.Join(dir, filename)

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, fmt.Sprintf("File stat Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := os.Remove(filePath); err != nil {
		http.Error(w, fmt.Sprintf("Delete Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
