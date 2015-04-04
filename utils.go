package main

import (
	"os"
)

func getListenAddress() string {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	// port defaults to 3000
	if port == "" {
		port = "8000"
	}
	return host + ":" + port
}
