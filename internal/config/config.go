package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFileName() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", homeDir, configFileName), nil
}

func Read() (*Config, error) {
	configFile, err := getConfigFileName()
	if err != nil {
		return nil, err
	}

	byteValue, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := json.Unmarshal(byteValue, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) SetUser(userName string) error {
	// structBytes, err := json.Marshal(*c)
	// if err != nil {
	// 	return err
	// }
	configFile, err := getConfigFileName()
	if err != nil {
		return err
	}
	c.CurrentUserName = userName

	file, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0775)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(*c); err != nil {
		return err
	}

	return nil
}
