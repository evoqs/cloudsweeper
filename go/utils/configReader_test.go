package utils

import (
	"fmt"
	"testing"
)

func TestHelloName(t *testing.T) {
	var cfg Config
	/*readFile(&cfg)
	readEnv(&cfg)
	fmt.Printf("%+v\n", cfg)*/
	cfg = LoadConfig()

	val, err := GetDBUrl(&cfg)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(val)
	}
}
