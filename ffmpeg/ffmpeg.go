package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsLocallyInstalled() bool {
	cmd := exec.Command("ffmpeg", "-version")

	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func ExtractAudio(videoFilePath, audioFilePath string) error {

	fmt.Printf("Extracting audio out of %v...\n", videoFilePath)

	cmd := exec.Command(
		"ffmpeg",
		"-i",
		videoFilePath,
		"-f",
		"mp3",
		"-ab",
		"19200",
		"-y",
		"-vn",
		audioFilePath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		fmt.Println("Error while extracting audio from video")
		fmt.Println(stderr.String())
		return err
	}

	fmt.Printf("Successfully extraced audio out of %v!\n", videoFilePath)
	return nil
}

func MergeSubtitle(subTitleFilePath, videoFilePath string) (string, error) {
	// outputFilePath := filepath.Join(filepath.Dir(videoFilePath), time.Now().String()+".mp4")
	outputFilePath := filepath.Join(filepath.Dir(videoFilePath), strings.TrimSuffix(filepath.Base(videoFilePath), filepath.Ext(videoFilePath))+"_output"+filepath.Ext(videoFilePath))
	cmd := exec.Command(
		"ffmpeg",
		"-i",
		videoFilePath,
		"-filter_complex",
		fmt.Sprintf("subtitles=%v", filepath.Base(subTitleFilePath)),
		outputFilePath,
	)

	cmd.Dir = filepath.Dir(subTitleFilePath)

	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		fmt.Println("Error while extracting audio from video")
		fmt.Println(stderr.String())
		return "", err
	}

	return outputFilePath, nil
}
