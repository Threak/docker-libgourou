package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func main() {
	// Define the directory where you want to store uploaded files
	uploadDir := "uploads"
	outputDir := "output"
	err := os.MkdirAll(uploadDir, 0755)
	if err != nil {
		fmt.Printf("Error creating upload directory: %v\n", err)
		return
	}

	// Define the upload handler function
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Parse the request with a maximum file size of 10 MB
			r.ParseMultipartForm(10 << 20)

			// Get the file from the request
			file, _, err := r.FormFile("file")
			if err != nil {
				fmt.Println("Error retrieving the file:", err)
				return
			}
			defer file.Close()

			// Create a new file in the upload directory with the same name as the uploaded file
			uploadPath := fmt.Sprintf("%s/%s", uploadDir, "contentMessage.acsm") // handler.Filename)
			f, err := os.Create(uploadPath)
			if err != nil {
				fmt.Println("Error creating the file:", err)
				return
			}
			defer f.Close()

			// Copy the uploaded file to the newly created file
			_, err = io.Copy(f, file)
			if err != nil {
				fmt.Println("Error copying the file:", err)
				return
			}

			commandName := "/usr/local/bin/acsmdownloader"
			commandArg  := []string{"--output-dir", outputDir, "--adept-directory", ".adept", uploadPath}

			command := exec.Command(commandName, commandArg...)
			output, err := command.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			pattern := regexp.MustCompile(`(.*)Created\s(.+)`)
			filenameMatch := pattern.FindSubmatch(output)
			filename := fmt.Sprintf("%s", filenameMatch[2])

			commandName = "/usr/local/bin/adept_remove"
			commandArg = []string{"--adept-directory", ".adept", filename}

			time.Sleep(1 * time.Second)

			command = exec.Command(commandName, commandArg...)
			_, err = command.Output()
			if err != nil {
				fmt.Println(err.Error())
			}

			err = os.Remove(uploadPath)
			if err != nil {
				fmt.Println("Error deleting the file:", err)
				return
			}

			fileBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			//w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/octet-stream")
			w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename[7:]))
			w.Header().Add("Content-Length", strconv.Itoa(len(fileBytes)))
			w.Write(fileBytes)
			return


		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve the static index.html file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(`<html>
<head><title>libgourou server</title></head>
<body>
<form action="/upload" method="post" enctype="multipart/form-data">
  <input type="file" name="file">
  <input type="submit" value="Upload">
</form>
</body>
</html>`)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Start the HTTP server
	port := "8080"
	fmt.Printf("Server is running on :%s...\n", port)
	http.ListenAndServe(":"+port, nil)
}
