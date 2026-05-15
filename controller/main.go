package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Set this to the IP of the machine running the worker
const workerIP = "http://10.179.171.130:8081"

func sendShutdown() {
	fmt.Println("Sending shutdown command...")
	resp, err := http.Get(workerIP + "/shutdown")
	if err != nil {
		fmt.Println("Connection Error:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Server response:", resp.Status)
}

func sendWallpaper(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening image file:", err)
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("wallpaper", filePath)
	if err != nil {
		fmt.Println("Error creating form:", err)
		return
	}
	io.Copy(part, file)
	writer.Close()

	fmt.Printf("Uploading %s to worker...\n", filePath)
	req, _ := http.NewRequest("POST", workerIP+"/wallpaper", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending wallpaper:", err)
		return
	}
	defer resp.Body.Close()

	msg, _ := io.ReadAll(resp.Body)
	fmt.Printf("Server says: %s\n", string(msg))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run controller.go shutdown")
		fmt.Println("  go run controller.go wallpaper <path_to_image.jpg>")
		return
	}

	command := os.Args[1]

	switch command {
	case "shutdown":
		sendShutdown()
	case "wallpaper":
		if len(os.Args) < 3 {
			fmt.Println("Please provide an image path.")
			return
		}
		sendWallpaper(os.Args[2])
	default:
		fmt.Println("Unknown command. Use 'shutdown' or 'wallpaper'.")
	}
}
