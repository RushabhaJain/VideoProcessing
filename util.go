package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cassava/lackey/audio/mp3"
)

func isFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getMP3FileDuration(filePath string) (time.Duration, error) {
	fileMetadata, err := mp3.ReadMetadata(filePath)

	if err != nil {
		fmt.Printf("Error while getting duration of mp3 file %v\n", filePath)
		fmt.Println(err)
		return 0, err
	}

	return fileMetadata.Length(), nil
}
