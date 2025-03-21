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

func writeConfigFile(config Config) error {

	filepath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	marshalledconfig, err := json.Marshal(config)
    if err != nil {
        return err
    }
	
	err = os.WriteFile(filepath, marshalledconfig, 0666)
	if err != nil {
		return fmt.Errorf("Error writing file:%s",err)
	}
	return nil
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


func (c Config)SetUser(username string) error {

	c.CurrentUser = username

	err := writeConfigFile(c)
	if err != nil {
		return err
	}
	
	return nil
}