package main

import (
	"encoding/json"
	"io/ioutil"
)

// Config represents the settings.json file.
type Config struct {
	// Twitter app configurations
	App struct {
		Key    string `json:"key"`
		Secret string `json:"secret"`
	} `json:"app"`
	// User specific configurations
	User struct {
		// This is the oauth token generated for this application
		Token string `json"token"`
		// This is the oauth screct that should be used by this
		// application.
		Secret string `json:"secret"`
	} `json:"user"`
}

// Read configuration file. Currently the configuration file is
// hardcoded as 'settings.json'. The following code is a configuration
// file example:
//
//      {
//          "app":
//          {
//              "key": "twitter-application-key",
//              "secret": "twitter-application-secret"
//          },
//          "user":
//          {
//              "token": "twitter-generated-token",
//              "secret": "twitter-generated-secret"
//          }
//      }
func readConfig() (*Config, error) {
	b, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = json.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
