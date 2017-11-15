package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type ConfigMail struct {
	From       string   `yaml:"from,omitempty"`
	FromName   string   `yaml:"from_name,omitempty"`
	Subject    string   `yaml:"subject,omitempty"`
	Recipients []string `yaml:"recipients,omitempty"`
}

type ConfigSlack struct {
	WebHookUrl string `yaml:"webhook_url,omitempty"`
	Channel    string `yaml:"channel,omitempty"`
	Author     string `yaml:"author,omitempty"`
	Message    string `yaml:"message,omitempty"`
	Username   string `yaml:"username,omitempty"`
	IconEmoji  string `yaml:"icon_emoji,omitempty"`
}

type Config struct {
	AllowedAddresses []string    `yaml:"allowed_addresses"`
	AllowedUsers     []string    `yaml:"allowed_users"`
	RecheckTime      int64       `yaml:"recheck_time"`
	GraceTime        int64       `yaml:"grace_time"`
	Mail             ConfigMail  `yaml:"mail,omitempty"`
	Slack            ConfigSlack `yaml:"slack,omitempty"`
}

func (config *Config) getConfig(configFile string) *Config {
	yamlFile, readErr := ioutil.ReadFile(configFile)
	if readErr != nil {
		log.Printf("yamlFile.Get err   #%v ", readErr)
		panic(readErr)
	}

	yamlErr := yaml.Unmarshal(yamlFile, &config)
	if yamlErr != nil {
		log.Fatalf("Unmarshal: %v", yamlErr)
		panic(yamlErr)
	}

	return config
}
