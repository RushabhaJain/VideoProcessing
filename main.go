package main

import "fmt"

func main() {
	fmt.Println("==================================")
	fmt.Println("Let's add subtitles to your videos")
	fmt.Println("==================================")

	// Take the source of videos
	fmt.Printf("Please add the path of folder where all videos exist: ")
	var videoFilePath string
	fmt.Scan(&videoFilePath)
}
