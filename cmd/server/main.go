package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/chameleon-c2/internal/crypto"
	"github.com/your-org/chameleon-c2/internal/server"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server server.Config `yaml:"server"`
}

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Generate master secret if not set
	if cfg.Server.MasterSecret == "" {
		secret, err := crypto.GenerateMasterSecret()
		if err != nil {
			log.Fatalf("failed to generate master secret: %v", err)
		}
		cfg.Server.MasterSecret = secret
		log.Printf("generated new master secret (save this!): %s", secret)
	}

	srv, err := server.New(&cfg.Server)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutting down...")
		srv.Stop()
		os.Exit(0)
	}()

	log.Fatal(srv.Start())
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: server.Config{
			ListenAddr:   ":443",
			DBPath:       "data/chameleon.db",
			OfflineAfter: 120,
			UseTLS:       false,
		},
	}
}
