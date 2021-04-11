package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// VendorConfig struct
type VendorConfig struct {
	Server struct {
		Host    string `yaml:"host"`
		Port    string `yaml:"port"`
		Options struct {
			MediaSuccessRatio float64  `yaml:"media_success_ratio"`
			MediaURLs         []string `yaml:"media_urls"`
			FirstMediaURL     string   `yaml:"first_media_url"`
			SecondMediaURL    string   `yaml:"second_media_url"`
		} `yaml:"options"`
	} `yaml:"server"`
}

// NewVendorConfig returns a struct composed of vendor config info
func NewVendorConfig(configPath string) (*VendorConfig, error) {

	config := &VendorConfig{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
