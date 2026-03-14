package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"Tella-Desktop/backend/utils/authutils"
	"Tella-Desktop/backend/utils/genericutil"
	"github.com/komkom/toml"
)

type Config struct {
	MaxFileSizeBytes int64 `json:"maxFileSizeBytes"`
	MaxFileCount     int   `json:"maxFileCount"`
	Port             int   `json:"defaultPort"`
}

var defaultMaxFileSize int64 = 3000000000 // 3 GB
var defaultMaxFileCount int = 1000
var defaultPort = 53320

func WriteDefaultConfig() {
	defaultConfig := fmt.Sprintf(`maxFileSizeBytes = %d
maxFileCount = %d
defaultPort = %d
`, defaultMaxFileSize, defaultMaxFileCount, defaultPort)
	err := os.WriteFile(authutils.GetConfigFilePath(), []byte(defaultConfig), genericutil.USER_ONLY_FILE_PERMS)
	if err != nil {
		panic(err)
	}
}

// TODO cblgh(2026-03-06): decide how to handle filesystem-level errors
func ReadConfig() Config {
	content, err := os.ReadFile(authutils.GetConfigFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			WriteDefaultConfig()
			return Config{defaultMaxFileSize, defaultMaxFileCount, defaultPort}
		} else {
			panic(err)
		}
	}
	var conf Config
	decoder := json.NewDecoder(toml.New(bytes.NewBuffer(content)))
	err = decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}
	return conf
}
