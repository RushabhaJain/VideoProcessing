package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/RushabhaJain/VideoProcessing/assemblyai"
	"github.com/RushabhaJain/VideoProcessing/ffmpeg"
	"github.com/cassava/lackey/audio/mp3"
	"github.com/joho/godotenv"
)

type fileToProcess struct {
	videoFilePath          string
	audioFilePath          string
	outputFilePath         string
	transcriptFilePath     string
	uploadedAudioUrl       string
	transcriptionProcessId string
}

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		panic("Missing .env file")
	}

	if os.Getenv("API_KEY") == "" {
		panic("Missing API_KEY in .env file")
	}

	if installed := ffmpeg.IsLocallyInstalled(); !installed {
		panic("Please install ffmpeg tool on your local machine")
	}

	API_KEY := os.Getenv("API_KEY")

	fmt.Println("==================================")
	fmt.Println("Let's add subtitles to your videos")
	fmt.Println("==================================")

	// Take the source of videos
	fmt.Printf("Please add the path of folder where all videos exist: ")
	var videoFileContainer string
	fmt.Scan(&videoFileContainer)

	filesToProcess := make([]fileToProcess, 0, 10)
	//TODO:: Validate videoFileContainer is valid path and is directory
	filepath.WalkDir(videoFileContainer, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".mp4" {
			filesToProcess = append(filesToProcess, fileToProcess{
				videoFilePath:      path,
				audioFilePath:      strings.TrimSuffix(path, filepath.Ext(path)) + "_audio.mp3",
				transcriptFilePath: strings.TrimSuffix(path, filepath.Ext(path)) + "_transcript.srt",
			})
		}
		return nil
	})

	// Waiting all goroutin to execute
	wg := sync.WaitGroup{}

	// Creating buffered channel for creating subtitles
	uploadAudioFileChannel := make(chan fileToProcess, 3)
	startTranscriptFileChannel := make(chan fileToProcess, 3)
	saveTranscriptFileChannel := make(chan fileToProcess, 3)
	mergeTranscriptFileChannel := make(chan fileToProcess, 3)

	go func(API_KEY string) {
		for videoFileToProcess := range uploadAudioFileChannel {
			go func(videoFileToProcess fileToProcess) {
				audioUrl, err := assemblyai.UploadFile(API_KEY, videoFileToProcess.audioFilePath)
				if err == nil {
					videoFileToProcess.uploadedAudioUrl = audioUrl
					startTranscriptFileChannel <- videoFileToProcess
				} else {
					wg.Done()
				}
			}(videoFileToProcess)
		}
	}(API_KEY)

	// Create worker to start transcription process
	go func(API_KEY string) {
		for videoFileToProcess := range startTranscriptFileChannel {
			go func(videoFileToProcess fileToProcess) {
				transcriptionProcessId, err := assemblyai.StartTranscriptProcess(API_KEY, videoFileToProcess.uploadedAudioUrl)
				if err == nil {
					videoFileToProcess.transcriptionProcessId = transcriptionProcessId

					duration, err := getMP3FileDuration(videoFileToProcess.audioFilePath)
					if err != nil {
						wg.Done()
						return
					}

					time.Sleep(time.Duration(duration.Seconds()*0.3) * time.Second)
					saveTranscriptFileChannel <- videoFileToProcess
				} else {
					wg.Done()
				}
			}(videoFileToProcess)
		}
	}(API_KEY)

	// Create worker to save transcripted file
	go func(API_KEY string) {
		for videoFileToProcess := range saveTranscriptFileChannel {
			go func(videoFileToProcess fileToProcess) {
				err := assemblyai.SaveTranscriptFile(API_KEY, videoFileToProcess.transcriptionProcessId, videoFileToProcess.transcriptFilePath)

				if err != nil {
					wg.Done()
				} else {
					mergeTranscriptFileChannel <- videoFileToProcess
				}
			}(videoFileToProcess)
		}
	}(API_KEY)

	// Create worker to save transcripted file
	go func() {
		for videoFileToProcess := range mergeTranscriptFileChannel {
			go func(videoFileToProcess fileToProcess) {
				outputFilePath, err := ffmpeg.MergeSubtitle(videoFileToProcess.transcriptFilePath, videoFileToProcess.videoFilePath)

				if err == nil {
					fmt.Printf("Subtitle for %v has merged into %v!\n", videoFileToProcess.videoFilePath, outputFilePath)
				}

				wg.Done()
			}(videoFileToProcess)
		}
	}()

	for _, videoFileToProcess := range filesToProcess {
		wg.Add(1)
		go func(videoFileToProcess fileToProcess) {
			err := ffmpeg.ExtractAudio(videoFileToProcess.videoFilePath, videoFileToProcess.audioFilePath)
			if err == nil {
				uploadAudioFileChannel <- videoFileToProcess
			}
		}(videoFileToProcess)
	}

	wg.Wait()
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
