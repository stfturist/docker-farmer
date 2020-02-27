package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

var c *Config

// Init will initialize the config file.
func Init(s string) {
	path := "config.json"

	if len(s) > 0 {
		path = s
	}

	file, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal(fmt.Sprintf("Config error: %v\n", err))
		return
	}

	var config *Config

	json.Unmarshal(file, &config)

	c = config

	if c.Title == "" {
		c.Title = "Farmer"
	}
}

// Config represents a config struct.
type Config struct {
	AllowedIPs []string `json:"allowed_ips"`
	Database   struct {
		Container string
		Password  string
		Prefix    string
		Type      string
		User      string
	}
	Domain string
	Docker struct {
		Host    string
		Version string
	}
	Logo       string
	Links      map[string]string
	Listen     string
	Menu       []*MenuItem
	Containers struct {
		Exclude []string
	}
	Title string
}

type MenuItem struct {
	Title string
	Link  string
}

// Get will return the config struct.
func Get() *Config {
	if c == nil {
		Init("")
	}

	return c
}
