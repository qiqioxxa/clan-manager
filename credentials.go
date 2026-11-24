package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func newClient(token string) (*CoCClient, error) {
	var developer *CoCDeveloperClient
	email, password, ok := loadDeveloperCredentials()
	if ok {
		developer = NewCoCDeveloperClient(email, password)
	}
	if token == "" && developer == nil {
		return nil, errors.New("no -token flag and no COC_DEV_EMAIL/COC_DEV_PASSWORD in .env")
	}

	client := NewCoCClient(developer, token)
	if token == "" {
		err := client.refreshToken("")
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func loadDeveloperCredentials() (email, password string, ok bool) {
	var paths []string
	exe, err := os.Executable()
	if err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), "..", ".env"))
	}
	paths = append(paths, ".env")

	for _, path := range paths {
		env, err := godotenv.Read(path)
		if err == nil && env["COC_DEV_EMAIL"] != "" && env["COC_DEV_PASSWORD"] != "" {
			return env["COC_DEV_EMAIL"], env["COC_DEV_PASSWORD"], true
		}
	}
	return "", "", false
}
