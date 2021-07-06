package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	rand "math/rand"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

func (config VendorConfig) mediaStatusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("mediaStatusHandler request obtained.")
	vars := mux.Vars(r)
	deviceId, ok := vars["deviceId"]
	if !ok {
		fmt.Println("deviceId is missing in parameters")
	}
	var status string

	successRatioThreshold := config.Server.Options.MediaStatusSuccessRatio
	if rand.Float64() <= successRatioThreshold {
		status = media_processing_workflow.Success
	} else {
		// return either `non_obtainable` or `pending` with equal probability
		if rand.Float64() <= 0.5 {
			status = media_processing_workflow.NotObtainable
		} else {
			status = media_processing_workflow.Pending
		}
	}
	mediaStatus := media_processing_workflow.MediaStatus{DeviceId: deviceId, Status: status}
	fmt.Println(fmt.Sprintf("%+v", mediaStatus))
	js, err := json.Marshal(mediaStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (config VendorConfig) mediaUrls(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceId, ok := vars["deviceId"]
	if !ok {
		fmt.Println("deviceId is missing in parameters")
	}

	mediaURLs := media_processing_workflow.MediaURLs{DeviceId: deviceId, Links: config.Server.Options.MediaURLs}
	js, err := json.Marshal(mediaURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (config VendorConfig) RunServer() {
	r := mux.NewRouter()
	r.HandleFunc("/mediastatus/{deviceId}", config.mediaStatusHandler)
	r.HandleFunc("/mediaurls/{deviceId}", config.mediaUrls)
	portAddress := fmt.Sprintf(":%s", config.Server.Port)

	// server for API endpoints that the workflow can utilize
	fmt.Println("Starting simulated vendor server...")
	_ = http.ListenAndServe(portAddress, r)
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
