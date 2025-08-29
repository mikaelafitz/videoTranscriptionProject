package main

import (
	"fmt"
	"os"
)

func main(){
	//check if filename is in the command line
	if len(os.Args) < 2 {
		fmt.Println("Please put file name in command line.")
		return
	}

	//access the video file
	videoFile := os.Args[1]

	//check if file is valid
	info, err := os.Stat(videoFile)
	if os.IsNotExist(err) {
		fmt.Println("file does not exist")
		return
	}
	if info.IsDir() {
		fmt.Println("this is a directory, not a file")
		return
	}
	fmt.Println("file is valid!")

	
}