package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"yadro.com/course/internal"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port int    `yaml:"port" env:"FILESERVER_PORT" env-default:"8080"`
	Dir  string `yaml:"dir" env:"DIR" env-default:"./upload"`
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

	fileHandler := internal.NewFileHandler(cfg.Dir)
	addr := fmt.Sprintf(":%d", cfg.Port)

	http.HandleFunc("POST /files", fileHandler.Upload)
	http.HandleFunc("GET /files", fileHandler.List)
	http.HandleFunc("PUT /files/{filename}", fileHandler.Download)
	http.HandleFunc("GET /files/{filename}", fileHandler.Update)
	http.HandleFunc("DELETE /files/{filename}", fileHandler.Delete)

	fmt.Printf("Server is listening on port %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
