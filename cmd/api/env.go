package main

import (
	"errors"
	"log"
	"os"
	"strconv"
)

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func getenv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	return v, nil
}
func getenvStr(key string) string {
	s, err := getenv(key)

	if err != nil {
		log.Fatal(err.Error())
	}

	return s
}

func getenvInt(key string) int {
	s, err := getenv(key)
	if err != nil {
		log.Fatal(err.Error())
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err.Error())
	}
	return v
}

func getenvBool(key string) bool {
	s, err := getenv(key)
	if err != nil {
		log.Fatal(err.Error())
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatal(err.Error())
	}
	return v
}
