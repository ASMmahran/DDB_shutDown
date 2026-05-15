package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// shutdownHandler executes the OS-specific shutdown command
func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received shutdown request...")
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("shutdown", "/s", "/t", "0")
	} else {
		cmd = exec.Command("shutdown", "-h", "now")
	}

	err := cmd.Run()
	if err != nil {
		http.Error(w, "Failed to shutdown: "+err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "Shutting down now...")
}

// wallpaperHandler receives an image and saves it locally
func wallpaperHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("wallpaper")
	if err != nil {
		http.Error(w, "Error retrieving file", 400)
		return
	}
	defer file.Close()

	// 1. الحصول على المسار الكامل وتجهيز اسم الملف
	pwd, _ := os.Getwd()
	imagePath := pwd + "\\new_wallpaper.jpg"

	// 2. إنشاء الملف وحفظه
	dst, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "Error saving file", 500)
		return
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		dst.Close()
		http.Error(w, "Error copying file", 500)
		return
	}

	// خطوة مهمة جداً: قفل الملف فوراً عشان ويندوز يقدر يقراه
	dst.Close()

	// 3. أمر PowerShell مع معالجة المسافات بشكل صحيح
	// نستخدم الحرف @ قبل النص في PowerShell للتعامل مع المسارات الصعبة
	psScript := fmt.Sprintf(`
		$Path = '%s'
		$code = @'
		using System.Runtime.InteropServices;
		public class Wallpaper {
			[DllImport("user32.dll", CharSet = CharSet.Auto)]
			public static extern int SystemParametersInfo(int uAction, int uParam, string lpvParam, int fuWinIni);
		}
'@
		Add-Type -TypeDefinition $code
		[Wallpaper]::SystemParametersInfo(20, 0, $Path, 3)
	`, imagePath)

	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("PowerShell Error: %s\n", string(output))
		http.Error(w, "Failed: "+err.Error(), 500)
		return
	}

	fmt.Println("Wallpaper applied successfully!")
	fmt.Fprintf(w, "Success!")
}
func main() {
	http.HandleFunc("/shutdown", shutdownHandler)
	http.HandleFunc("/wallpaper", wallpaperHandler)

	port := ":8081"
	fmt.Printf("Worker Node active. Listening on %s...\n", port)

	// Start server
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
