package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {

	listenAddress := "0.0.0.0:80"
	Yt3Domain := "https://yt3.ggpht.com"
	imageFolder := "/yt3"
	contentFolder := "/content"

	metrics := NewMetrics()

	http.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		fmt.Fprint(w, "ok")
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		metrics.ServeHTTP(w, r)
	})

	http.HandleFunc("/yt3/", metrics.WithMetrics(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

		u, _ := url.Parse(r.URL.String())
		req := u.Path[6:]

		imgId := strings.Split(req, "=")[0]
		imgVer := strings.Split(req, "=")[1]

		imagePath := imageFolder + imgId + "/" + imgVer + ".png"

		if _, err := os.Stat(imagePath); err == nil {
			// Image exists, serve it
			http.ServeFile(w, r, imagePath)
			return
		}

		imageURL := Yt3Domain + req

		log.Println("Fetching image from", imageURL)

		resp, err := http.Get(imageURL)
		if err != nil {
			log.Println("Failed to fetch image1", err)
			http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Println("Failed to fetch image2", err)
			http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
			return
		}

		// Create the images folder if it doesn't exist
		if err := os.MkdirAll(imageFolder+"/"+imgId, os.ModePerm); err != nil {
			log.Println("Failed to create image folder", err)
			http.Error(w, "Failed to create image folder", http.StatusInternalServerError)
			return
		}

		// Save the image to the folder
		out, err := os.Create(imagePath)
		if err != nil {
			log.Println("Failed to save image", err)
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Println("Failed to save image", err)
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}

		// Serve the saved image
		http.ServeFile(w, r, imagePath)
	}))

	http.HandleFunc("/file/", metrics.WithMetrics(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

		u, _ := url.Parse(r.URL.String())
		req := u.Path[5:]

		filePath := contentFolder + req

		log.Printf("Serving file %s", filePath)

		if _, err := os.Stat(filePath); err == nil {
			// File exists, serve it
			http.ServeFile(w, r, filePath)
			return
		}

		http.Error(w, "File not found", http.StatusNotFound)
	}))

	log.Println("HTTP webserver running. Access it at", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
