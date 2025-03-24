package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const (
	MACOS_LOCATION   = "~/.config/" + APP_NAME + "/config.json"
	WINDOWS_LOCATION = "C:\\Users\\%USERNAME%\\AppData\\Roaming\\" + APP_NAME + "\\config.json"
	LINUX_LOCATION   = "~/.config/" + APP_NAME + "/config.json"
)

// Configuration Structs for the configuration
type Configuration struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Configurations struct {
	Configurations []Configuration `json:"configurations"`
}

// ConfigurationFileExists checks if the configuration file exists
func GetConfigLocation() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting user home directory: %v", err)
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, ".config", APP_NAME, "config.json")
	case "windows":
		return filepath.Join(homeDir, "AppData", "Roaming", APP_NAME, "config.json")
	case "linux":
		return filepath.Join(homeDir, ".config", APP_NAME, "config.json")
	default:
		return filepath.Join(homeDir, ".config", APP_NAME, "config.json")
	}
}

func ConfigurationFileExists() bool {
	configPath := GetConfigLocation()
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return false
	}

	_, err = os.Stat(absPath)
	exists := !os.IsNotExist(err)

	//if exists {
	//	fmt.Printf("Configuration file found\n")
	//} else {
	//	fmt.Printf("Configuration file not found\n")
	//}

	return exists
}

func getDefaultPath() string {
	if os.Getenv("GOOS") == "windows" {
		return "C:\\Users\\%USERNAME%\\" + APP_NAME
	} else if os.Getenv("GOOS") == "darwin" {
		return "/Users/%USERNAME%/" + APP_NAME
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		return home + "/" + APP_NAME
	}
}

func ensureConfigFile() (*os.File, error) {
	configPath := GetConfigLocation()

	// Create all parent directories with proper permissions
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create or open the config file
	file, err := os.Create(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %w", err)
	}

	return file, nil
}

// CreateConfigurationFile creates the configuration file
func CreateConfigurationFile() Configurations {
	configurations := Configurations{
		Configurations: []Configuration{
			{
				Key:   "interval",
				Value: "5",
			},
			{
				Key:   "target",
				Value: "local",
			},
			{
				Key:   "local_path",
				Value: getDefaultPath(),
			},
			{
				Key:   "local_file_retention",
				Value: "20000",
			},
		},
	}
	file, err := ensureConfigFile()
	if err != nil {
		panic(err)
	}
	err = json.NewEncoder(file).Encode(configurations)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	return configurations
}

// ReadConfigurationFile reads the configuration file
func ReadConfigurationFile() (Configurations, error) {
	var configurations Configurations

	configPath := GetConfigLocation()
	if configPath == "" {
		return configurations, fmt.Errorf("failed to get config location")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return configurations, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: failed to close config file: %v", closeErr)
		}
	}()

	if err := json.NewDecoder(file).Decode(&configurations); err != nil {
		return configurations, fmt.Errorf("failed to decode config file: %w", err)
	}

	return configurations, nil
}

// WriteConfigurationFile writes the configuration file
func WriteConfigurationFile(configurations Configurations) error {
	if !ConfigurationFileExists() {
		fmt.Println("Configuration file does not exist. Creating a new one...")
		CreateConfigurationFile()
	} else {
		fmt.Println("Configuration file exists. Updating the existing one...")
	}
	file, err := os.Create(GetConfigLocation())
	if err != nil {
		return err
	}
	defer func(file *os.File) error {
		err := file.Close()
		if err != nil {
			return err
		}
		return nil
	}(file)

	err = json.NewEncoder(file).Encode(configurations)
	if err != nil {
		return err
	}
	return nil
}

func InitConfig() Configurations {
	if !ConfigurationFileExists() {
		return CreateConfigurationFile()
	}
	return Configurations{}
}

func GetConfig() Configurations {
	config, err := ReadConfigurationFile()
	if err != nil {
		log.Printf("Error reading configuration file: %v", err)
		return InitConfig()
	}
	return config
}

func GetTmpDir() string {
	tmpDir := os.TempDir() + APP_NAME
	return tmpDir
}

func GetConfigArray() map[string]string {
	configs := GetConfig()
	result := make(map[string]string)

	for _, config := range configs.Configurations {
		result[config.Key] = config.Value
	}

	return result
}
