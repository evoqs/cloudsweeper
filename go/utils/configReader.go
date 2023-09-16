package utils

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

var conf Config

type Config struct {
	Server struct {
		Port string `yaml:"port" envconfig:"SERVER_PORT"`
		Host string `yaml:"host" envconfig:"SERVER_HOST"`
	} `yaml:"server"`
	Database struct {
		Username string `yaml:"user" envconfig:"DB_USERNAME"`
		Password string `yaml:"pass" envconfig:"DB_PASS"`
		Host     string `yaml:"host" envconfig:"DB_HOST"`
		Port     string `yaml:"port" envconfig:"DB_PORT"`
		Name     string `yaml:"dbname" envconfig:"DB_NAME"`
		Url      string //URL is constructed from given params
	} `yaml:"database"`
	Custodian struct {
		C7nAwsInstall string `yaml:"c7nawsinstall" envconfig:"C7N_AWS_INSTALL"`
	} `yaml:"custodian"`
}

// Loading the cfg
func LoadConfig() Config {

	readFile(&conf)
	readEnv(&conf)
	return conf
}

func GetConfig() Config {
	return conf
}

func readFile(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}
