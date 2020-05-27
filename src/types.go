package main

import "github.com/gorilla/mux"

type RegisterUser struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	Email          string `json:"email"`
	InsitutionName string `json:"institutionname"`
}

type IdentityServerResponse struct {
	ID       string `json:"id"`
	Username string `json:"userName"`
}

type RegisterUserResponse struct {
	UserCreated string `json:"usercreated"`
	Username    string `json:"username"`
	UserID      string `json:"id"`
	Message     string `json:"message"`
}

type Server struct {
	router *mux.Router
}
type Config struct {
	IS_Host         string
	IS_Port         string
	ListenServePort string
	IS_Username     string
	IS_Password     string
	UM_Host         string
	UM_Port         string
}
