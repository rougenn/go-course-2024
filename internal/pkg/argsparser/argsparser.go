package argsparser

import (
	"fmt"
	"os"
	"strconv"
)

func ParseArgs() (int, int, string, string) {
	saveDuration, ok := os.LookupEnv("SAVE_DURATION")
	if !ok {
		fmt.Println("save dur is not provided")
		saveDuration = "60"
	}
	clearDuration, ok := os.LookupEnv("CLEAR_DURATION")
	if !ok {
		fmt.Println("clear dur is not provided")
		saveDuration = "60"
	}
	filename, ok := os.LookupEnv("STORAGE_FILENAME")
	if !ok {
		fmt.Println("save dir is not provided")
		filename = "my-storage.json"
	}
	port, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		fmt.Println("port is not provided")
		port = "8090"
	}

	SD, err := strconv.Atoi(saveDuration)
	if err != nil {
		fmt.Println("incorrect format of saveDuration, set to default")
		SD = 60
	}

	CD, err := strconv.Atoi(clearDuration)

	if err != nil {
		fmt.Println("incorrect format of saveDuration, set to default")
		CD = 60
	}

	return SD, CD, filename, ":" + port
}
