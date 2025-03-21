package config
import (
	"encoding/json"
	"os"
	"fmt"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DatabaseURL string `json:"db_url"`
	CurrentUser string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {

	homefolder, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to access home folder: %s", err)
	}

	return fmt.Sprintf("%s/%s", homefolder, configFileName), nil
}

func Read() (Config, error) {

	filepath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	body, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, fmt.Errorf("Error reading file: %s", err)
	}
	if len(body) == 0 {
		return Config{}, fmt.Errorf("0 bytes read from config file")
	}

	config := Config{}
	err = json.Unmarshal(body, &config)
	if err != nil {
		return Config{}, fmt.Errorf("config file error: %s", err)
	}
	return config, nil
}


func (Config)SetUser(username string) {

}