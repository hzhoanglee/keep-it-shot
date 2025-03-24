package main

import (
	"fmt"
	"github.com/kbinani/screenshot"
	"github.com/skratchdot/open-golang/open"
	"image"
	"image/png"
	"os"
	"sort"
	"strconv"
	"time"
)

func StartService() {
	fmt.Printf("Starting service\n")
	waitTime, _ := strconv.Atoi(config["interval"])

	stopChan := make(chan struct{})

	SetStopChannel(stopChan)

	go func(waitTime int, stop <-chan struct{}) {
		for {
			select {
			case <-stop:
				fmt.Println("Screen capture goroutine stopped")
				return
			default:
				handleCaptureToLocal()
				waitDuration := time.Duration(waitTime) * time.Second
				time.Sleep(waitDuration)
			}
		}
	}(waitTime, stopChan)

	fmt.Printf("Waiting %s seconds\n", waitTime)
}

var stopChannel chan struct{}

func SetStopChannel(ch chan struct{}) {
	stopChannel = ch
}

func StopService() {
	fmt.Printf("Stopping service\n")

	if stopChannel != nil {
		close(stopChannel)
		stopChannel = nil
	}

}

func handleCaptureToLocal() {
	fmt.Println("Capturing to local")
	imgList := captureScreenToTmp()
	if imgList == nil {
		return
	}

	for i, img := range imgList {
		fmt.Println("Saving capture ", i)
		dateTimeNow := time.Now().Format("2006-01-02 15-04-05")
		fileName := fmt.Sprintf("%s/%s_%s_%d.png", config["local_path"], dateTimeNow, APP_NAME, i)
		err := os.MkdirAll(config["local_path"], os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory")
			return
		}
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Error creating file")
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println("Error closing file")
			}
		}(file)
		_ = png.Encode(file, img)
		fmt.Println("Capture saved to ", fileName)
	}

}

func captureScreenToTmp() []image.Image {
	n := screenshot.NumActiveDisplays()
	imgList := make([]image.Image, n)
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			fmt.Println("Failed to capture screen")
			return nil
		}
		imgList[i] = img
	}
	return imgList
}

func checkForRetentionPolicy() {
	maxFiles := config["local_file_retention"]
	if maxFiles == "" {
		return
	}
	maxFilesInt, err := strconv.Atoi(maxFiles)
	if err != nil {
		fmt.Println("Error converting local_file_retention to int")
		return
	}
	//files, err := os.ReadDir(config["local_path"])
	files, err := os.ReadDir(config["local_path"])
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	if len(files) <= maxFilesInt {
		fmt.Printf("Number of files (%d) is less than or equal to the retention policy (%d)\n", len(files), maxFilesInt)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		infoI, errI := files[i].Info()
		infoJ, errJ := files[j].Info()
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	for i := 0; i < len(files)-maxFilesInt; i++ {
		fmt.Println("Removing file:", files[i].Name())
		err := os.Remove(config["local_path"] + "/" + files[i].Name())
		if err != nil {
			fmt.Println("Error removing file:", err)
		}
	}

}

func QuitApp() {
	StopService()
	os.Exit(0)
}

func OpenLocalFolder() {
	err := open.Start(config["local_path"])
	if err != nil {
		fmt.Println("Error opening local folder:", err)
	}
}
