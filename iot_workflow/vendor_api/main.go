package main

import (
	"flag"
	"fmt"
	"log"
	rand "math/rand"
	"net/http"
	"os"
)

// ValidateConfigPath ..
func ValidateConfigPath(configPath string) error {
	s, err := os.Stat(configPath)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' file is not a regular file; please verify", configPath)
	}
	return nil
}

func parseFlags() (string, error) {
	var configPath string

	flag.StringVar(&configPath, "config", "./config.yaml", "the path to the vendor config file. Defaults to the config.yaml file")

	flag.Parse()

	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}

func (config VendorConfig) mediaStatusHandler(w http.ResponseWriter, _ *http.Request) {
	successRatioThreshold := config.Server.Options.MediaSuccessRatio
	if rand.Float64() <= successRatioThreshold {
		fmt.Fprintf(w, "success")
	} else {
		fmt.Fprintf(w, "pending")
	}
}

func (config VendorConfig) firstMediaLinkHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, config.Server.Options.FirstMediaURL)
}

func (config VendorConfig) secondMediaLinkHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, config.Server.Options.SecondMediaURL)
}

func (config VendorConfig) RunServer() {

	http.HandleFunc("/mediastatus", config.mediaStatusHandler)
	http.HandleFunc("/firstmedialink", config.firstMediaLinkHandler)
	http.HandleFunc("/secondmedialink", config.secondMediaLinkHandler)
	portAddress := fmt.Sprintf(":%s", config.Server.Port)

	// server for API endpoints that the workflow can utilize
	fmt.Println("Starting simulated vendor server...")
	_ = http.ListenAndServe(portAddress, nil)
}

func main() {

	vendorCfgPath, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := NewVendorConfig(vendorCfgPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg.RunServer()
}
