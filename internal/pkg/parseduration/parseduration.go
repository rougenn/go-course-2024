package parseduration

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultSaveDuration  = 60
	defaultFilename      = "my-storage.json"
	defaultPort          = "8090"
	defaultClearDuration = 60
)

func ParseDuration() (int, int, string, string) {
	saveDuration, ok := os.LookupEnv("SAVE_DURATION")
	if !ok {
		fmt.Println("save dur is not provided")
	}
	clearDuration, ok := os.LookupEnv("CLEAR_DURATION")
	if !ok {
		fmt.Println("clear dur is not provided")
	}
	filename, ok := os.LookupEnv("STORAGE_FILENAME")
	if !ok {
		fmt.Println("save dir is not provided")
		filename = defaultFilename
	}
	port, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		fmt.Println("port is not provided")
		port = defaultPort
	}

	SD, err := strconv.Atoi(saveDuration)
	if err != nil {
		fmt.Println("incorrect format of saveDuration, set to default")
		SD = defaultSaveDuration
	}

	CD, err := strconv.Atoi(clearDuration)

	if err != nil {
		fmt.Println("incorrect format of saveDuration, set to default")
		CD = defaultClearDuration
	}

	return SD, CD, filename, ":" + port
}
