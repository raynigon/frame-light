package controller

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

//go:embed ui/*
var uiFiles embed.FS

// fileStat retrieves the file information for embedded files
func fileStat(file fs.File) fs.FileInfo {
	info, _ := file.Stat()
	return info
}

// UIHandler serves files from the embedded UI folder
func UIHandler(w http.ResponseWriter, r *http.Request) {
	// The path contains the prefix "/ui", so we need to strip it
	filePath := strings.TrimPrefix(r.URL.Path, "/ui")
	// Remove any leading slashes, to avoid directory traversal attacks
	filePath = strings.TrimPrefix(filePath, "/")
	// Check if the path starts with a dot or contains "..", if so reject the request with 400
	if strings.HasPrefix(filePath, ".") || strings.Contains(filePath, "..") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// Check if the path is empty, if so serve the index.html file
	if filePath == "" {
		filePath = "index.html"
	}
	// Open the file from the embedded filesystem
	file, err := uiFiles.Open("ui/" + filePath)
	if err != nil {
		log.Warnf("Error opening file: %v", err)
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// Serve the file
	http.ServeContent(w, r, filePath, fileStat(file).ModTime(), file.(io.ReadSeeker))
}
