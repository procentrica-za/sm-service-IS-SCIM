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
	fmt.Printf("IS_Host: %v\n", config.IS_Host)
	fmt.Printf("IS_Port: %v\n", config.IS_Port)
	fmt.Printf("Listening and Serving on Port: %v\n", config.ListenServePort)
	fmt.Printf("UM_Host: %v\n", config.UM_Host)
	fmt.Printf("UM_Port: %v\n", config.UM_Port)
}

func CreateConfig() Config {
	conf := Config{
		IS_Host:         os.Getenv("IS_HOST"),
		IS_Port:         os.Getenv("IS_PORT"),
		ListenServePort: os.Getenv("LISTEN_AND_SERVE_PORT"),
		IS_Username:     os.Getenv("IS_USERNAME"),
		IS_Password:     os.Getenv("IS_PASSWORD"),
		UM_Host:         os.Getenv("UM_HOST"),
		UM_Port:         os.Getenv("UM_PORT"),
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
