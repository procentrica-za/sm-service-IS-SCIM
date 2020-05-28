package main

import "github.com/gorilla/mux"

type RegisterUser struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	Email          string `json:"email"`
	InsitutionName string `json:"institutionname"`
	KeySecret      string `json:"keysecret"`
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

type UpdateUser struct {
	UserID         string `json:"id"`
	ScimID         string `json:"scimid"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	Email          string `json:"email"`
	InsitutionName string `json:"institutionname"`
}

type UpdateUserDB struct {
	UserID         string `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	Email          string `json:"email"`
	InsitutionName string `json:"institutionname"`
}

type UpdateUserResult struct {
	UserUpdated bool   `json:"userupdated"`
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
	Key_Secret      string
}
