package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

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

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)

	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/hello", helloHandler)

	fmt.Printf("Server is listening on port %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %s", err.Error())
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		log.Printf("Write Error: %s", err.Error())
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, "Hello, %s!", name); err != nil {
		log.Printf("Write Error: %s", err.Error())
	}
}
