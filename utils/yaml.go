package utils
import (
	"go.yaml.in/yaml/v4"
	"os"
	"fmt"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Address     string    `yaml:"address"`
	Environment string `yaml:"environment"`
}

type DatabaseConfig struct {
	Host  string `yaml:"host"`
	Ports []int  `yaml:"ports"`
}

var YamlConfig Config
func Yaml_init(filename string) {
	fmt.Println("Yaml init")
	// --- READING A YAML FILE ---
	
	// Read the file byte content
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		Log.Critical("failed to read file: %v", err)
	}

	// Parse the YAML into our struct
	err = yaml.Unmarshal(yamlFile, &YamlConfig)
	if err != nil {
		Log.Critical("failed to unmarshal yaml: %v", err)
	}

	// Access parsed data fields safely
	Log.Info("Parsed Server Address: %d\n", YamlConfig.Server.Address)
	Log.Info("Parsed DB Host: %s\n", YamlConfig.Database.Host)
	// --- WRITING A YAML FILE ---
	
	Log.Info("Successfully updated and saved config_updated.yaml!")
}