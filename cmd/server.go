package main

import "net/http"

func main() {
	srv := http.NewServeMux()

	// serve files from the static directory
	srv.Handle("/", http.FileServer(http.Dir("static")))

	println("Server started at http://localhost:8080/")
	http.ListenAndServe(":8080", srv)
}
