package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var config Config

func init() {
	config = CreateConfig()
	fmt.Printf("IS_Host: %v\n", config.ISHost)
	fmt.Printf("IS_Port: %v\n", config.ISPort)
	fmt.Printf("APIM_Host: %v\n", config.APIMHost)
	fmt.Printf("APIM_Port: %v\n", config.APIMPort)
	fmt.Printf("Listening and Serving on Port: %v\n", config.ListenServePort)
	fmt.Printf("UM_Host: %v\n", config.UMHost)
	fmt.Printf("UM_Port: %v\n", config.UMPort)
}

func CreateConfig() Config {
	conf := Config{
		ISHost:          os.Getenv("IS_HOST"),
		ISPort:          os.Getenv("IS_PORT"),
		APIMHost:        os.Getenv("APIM_HOST"),
		APIMPort:        os.Getenv("APIM_PORT"),
		ListenServePort: os.Getenv("LISTEN_AND_SERVE_PORT"),
		ISUsername:      os.Getenv("IS_USERNAME"),
		ISPassword:      os.Getenv("IS_PASSWORD"),
		UMHost:          os.Getenv("UM_HOST"),
		UMPort:          os.Getenv("UM_PORT"),
		KeySecret:       os.Getenv("KEY_SECRET"),
	}
	return conf
}

func main() {
	server := Server{
		router: mux.NewRouter(),
	}
	//Set up routes for server
	server.routes()
	handler := removeTrailingSlash(server.router)
	fmt.Printf("starting server on port " + config.ListenServePort + "...\n")
	log.Fatal(http.ListenAndServe(":"+config.ListenServePort, handler))
}
func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}
