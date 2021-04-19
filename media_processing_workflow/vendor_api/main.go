package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	rand "math/rand"
	"net/http"
	"os"

	"github.com/nirpadma/temporal-workflows/media_processing_workflow"
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
	successRatioThreshold := config.Server.Options.MediaStatusSuccessRatio
	if rand.Float64() <= successRatioThreshold {
		fmt.Fprintf(w, media_processing_workflow.Success)
	} else {
		// return either `non_obtainable` or `pending` with equal probability
		if rand.Float64() <= 0.5 {
			fmt.Fprintf(w, media_processing_workflow.NotObtainable)
		} else {
			fmt.Fprintf(w, media_processing_workflow.Pending)
		}

	}
}

func (config VendorConfig) mediaUrls(w http.ResponseWriter, _ *http.Request) {

	mediaURLs := media_processing_workflow.MediaURLs{Links: config.Server.Options.MediaURLs}
	js, err := json.Marshal(mediaURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (config VendorConfig) RunServer() {

	http.HandleFunc("/mediastatus", config.mediaStatusHandler)
	http.HandleFunc("/mediaurls", config.mediaUrls)
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
