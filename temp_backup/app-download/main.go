package main

import (
	"fmt"
	"net/http"

	"github.com/JackovAlltrades/go-toolbox"
)

func main() {
	// get routes
	mux := routes()

	// start a server
	fmt.Println("Server started at http://localhost:8090")
	err := http.ListenAndServe(":8090", mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/download", downloadFile)

	return mux
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	t := toolbox.Tools{}
	t.DownloadStaticFile(w, r, "./files", "pic.png", "pic2.png")
}
