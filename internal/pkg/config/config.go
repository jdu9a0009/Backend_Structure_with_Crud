package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	ErrorBotToken string   `json:"ERROR_BOT_TOKEN" yaml:"ERROR_BOT_TOKEN"`
	ErrorChatID   []string `json:"ERROR_CHAT_ID" yaml:"ERROR_CHAT_ID"`
}

func NewConfig() *Config {
	var c *Config

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err #%v", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
