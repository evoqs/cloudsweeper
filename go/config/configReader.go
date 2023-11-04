package config

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
	Aws struct {
		Creds struct {
			Aws_access_key_id     string `yaml:"aws_access_key_id" envconfig:"AWS_ACCESS_KEY_ID"`
			Aws_default_region    string `yaml:"aws_default_region" envconfig:"AWS_DEFAULT_REGION"`
			Aws_secret_access_key string `yaml:"aws_secret_access_key" envconfig:"AWS_SECRET_ACCESS_KEY"`
		} `yaml:"creds"`
	} `yaml:"aws"`
	Logging struct {
		Default_log_file string `yaml:"default_log_file" envconfig:"DEFAULT_LOG_FILE"`
		Max_size         int    `yaml:"max_size" envconfig:"MAX_SIZE"`
		Max_backups      int    `yaml:"max_backups" envconfig:"MAX_BACKUPS"`
		Max_age          int    `yaml:"max_age" envconfig:"MAX_AGE"`
	} `yaml:"logging"`
	OpenAI struct {
		ApiKey string `yaml:"openai_apikey" envconfig:"OPENAI_APIKEY"`
		Url    string `yaml:"openai_url" envconfig:"OPENAI_URL"`
	} `yaml:"openai"`
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
