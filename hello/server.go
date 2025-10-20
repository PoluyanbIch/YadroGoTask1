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

	fmt.Printf("Server is listening on port %d...\n", addr)
	http.ListenAndServe(addr, nil)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hello, %s!", name)))
}
