package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port       int    `json:"port"`
	Password   string `json:"password"`
	SessionKey string `json:"session_key"`
	DBPath     string `json:"db_path"`
	BaseURL    string `json:"base_url"`
	AppName    string `json:"app_name"`
}

var AppConfig Config

func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		// 使用默认配置
		AppConfig = Config{
			Port:       8080,
			Password:   "nibstash",
			SessionKey: "nibstash-secret-key-change-me",
			DBPath:     "data/nibstash.db",
			BaseURL:    "http://localhost:8080",
			AppName:    "囤囤鼠",
		}
		return SaveConfig(path)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		return err
	}
	return nil
}

func SaveConfig(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(AppConfig)
}
