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
	WebhookUrl string `yaml:"webhook_url,omitempty"`
	Channel    string `yaml:"channel,omitempty"`
	Author     string `yaml:"author,omitempty"`
	Message    string `yaml:"message,omitempty"`
	Username   string `yaml:"username,omitempty"`
	IconEmoji  string `yaml:"icon_emoji,omitempty"`
}

type Config struct {
	AllowedAddresses []string    `yaml:"allowed_addresses,omitempty"`
	AllowedUsers     []string    `yaml:"allowed_users,omitempty"`
	RecheckTime      int64       `yaml:"recheck_time,omitempty"`
	GraceTime        int64       `yaml:"grace_time,omitempty"`
	Mail             ConfigMail  `yaml:"mail,omitempty"`
	Slack            ConfigSlack `yaml:"slack,omitempty"`
}

func (c *Config) getConfig(configfile string) *Config {
	yamlFile, readErr := ioutil.ReadFile(configfile)
	if readErr != nil {
		log.Printf("yamlFile.Get err   #%v ", readErr)
		panic(readErr)
	}

	yamlErr := yaml.Unmarshal(yamlFile, &c)
	if yamlErr != nil {
		log.Fatalf("Unmarshal: %v", yamlErr)
		panic(yamlErr)
	}

	return c
}
