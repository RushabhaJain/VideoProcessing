package assemblyai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func UploadFile(apiKey, audioFilePath string) (string, error) {

	fmt.Printf("Uploading file %v...\n", audioFilePath)

	const UPLOAD_URL = "https://api.assemblyai.com/v2/upload"

	content, err := os.ReadFile(audioFilePath)

	if err != nil {
		fmt.Printf("Error in uploading file %v!", audioFilePath)
		return "", err
	}

	client := http.Client{}
	req, _ := http.NewRequest("POST", UPLOAD_URL, bytes.NewBuffer(content))
	req.Header.Set("authorization", apiKey)
	res, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error in uploading file %v!", audioFilePath)
		return "", err
	}
	defer res.Body.Close()
	fmt.Printf("Uploaded file %v!\n", audioFilePath)

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)
	return result["upload_url"].(string), nil
}

func StartTranscriptProcess(apiKey, audioFileUrl string) (string, error) {
	const TRANSCRIPT_URL = "https://api.assemblyai.com/v2/transcript"
	fmt.Printf("Started transcripting of %v...\n", audioFileUrl)

	client := http.Client{}

	byteArray, err := json.Marshal(map[string]string{
		"audio_url": audioFileUrl,
	})

	if err != nil {
		fmt.Printf("Error in transcripting audio file %v!\n", audioFileUrl)
		return "", err
	}

	req, _ := http.NewRequest("POST", TRANSCRIPT_URL, bytes.NewBuffer(byteArray))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", apiKey)
	res, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error in transcripting audio file %v!\n", audioFileUrl)
		return "", err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)
	return result["id"].(string), nil
}

func SaveTranscriptFile(apiKey, transcriptProcessId, filePath string) error {
	SAVE_TRANSCRIPT_URL := fmt.Sprintf("https://api.assemblyai.com/v2/transcript/%v/srt", transcriptProcessId)

	fmt.Printf("Saving transcript file %v...\n", filePath)

	client := http.Client{}

	req, _ := http.NewRequest("GET", SAVE_TRANSCRIPT_URL, nil)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", apiKey)
	res, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error in saving transcript file %v...\n", filePath)
		return err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	os.WriteFile(filePath, data, 0644)
	return nil
}
