package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Range")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// List 
func listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var movies []string
	err := filepath.Walk("./movies", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			movies = append(movies, info.Name())
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to list movies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, movie := range movies {
		fmt.Fprintf(w, "%s\n", movie)
	}
}

// Stream the movie
func streamMovieHandler(w http.ResponseWriter, r *http.Request) {
	movieName := r.URL.Query().Get("name")
	if movieName == "" {
		http.Error(w, "Movie name is required", http.StatusBadRequest)
		return
	}

	moviePath := filepath.Join("./movies", movieName)
	file, err := os.Open(moviePath)
	if err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get the file info for the content length
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Unable to get file info", http.StatusInternalServerError)
		return
	}

	fileSize := fileInfo.Size()
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")

	// Handle range requests
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		// Parse the range header
		var start, end int64
		fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
		if end == 0 {
			end = fileSize - 1
		}

		// Set the appropriate headers
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
		w.WriteHeader(http.StatusPartialContent)

		// Stream the requested range
		file.Seek(start, 0)
		buffer := make([]byte, 1024*8) // 8KB buffer
		for {
			if start >= end {
				break
			}
			n, err := file.Read(buffer)
			if err != nil {
				break
			}
			w.Write(buffer[:n])
			start += int64(n)
		}
	} else {
		// If no range is specified, send the whole file
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		http.ServeFile(w, r, moviePath)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/api/movies", enableCORS(http.HandlerFunc(listMoviesHandler)))
	mux.Handle("/api/stream", enableCORS(http.HandlerFunc(streamMovieHandler)))

	fmt.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
