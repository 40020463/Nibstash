package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port       int    `json:"port"`
	Password   string `json:"password"`
	JWTSecret  string `json:"jwt_secret"`
	DBPath     string `json:"db_path"`
	BaseURL    string `json:"base_url"`
	AppName    string `json:"app_name"`
	EncryptKey string `json:"encrypt_key"` // AES-GCM 加密密钥 (32字节)
}

var App Config

func Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		// 使用默认配置
		App = Config{
			Port:       8080,
			Password:   "nibstash",
			JWTSecret:  "nibstash-jwt-secret-change-me-32bytes!",
			DBPath:     "data/nibstash.db",
			BaseURL:    "http://localhost:8080",
			AppName:    "囤囤鼠",
			EncryptKey: "nibstash-encrypt-key-32-bytes!!!", // 32字节
		}
		return Save(path)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&App)
	if err != nil {
		return err
	}
	return nil
}

func Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(App)
}
