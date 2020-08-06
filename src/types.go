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

type TokenResponse struct {
	Accesstoken  string `json:"access_token"`
	Refreshtoken string `json:"refresh_token"`
}

type LoginUserResult struct {
	UserID       string `json:"id"`
	Username     string `json:"username"`
	UserLoggedIn bool   `json:"userloggedin"`
	Institution  string `json:"institution"`
	Message      string `json:"message"`
	Accesstoken  string `json:"access_token"`
	Refreshtoken string `json:"refresh_token"`
}

type UpdatePassword struct {
	UserID          string `json:"id"`
	ScimID          string `json:"scimid"`
	Username        string `json:"username"`
	CurrentPassword string `json:"currentpassword"`
	Password        string `json:"password"`
}

type UpdatePasswordResult struct {
	PasswordUpdated bool   `json:"passwordupdated"`
	Message         string `json:"message"`
}

type ScimID struct {
	ScimID string `json:"id"`
}

type getPassword struct {
	UserID   string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	GotUser  bool   `json:"gotuser"`
}

type EmailResult struct {
	Message  string `json:"message"`
	Password string `json:"password"`
}

type UserResult struct {
	Message string `json:"message"`
}

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Resources struct {
	ScimID string `json:"id"`
}

type IdentityServerSCIMID struct {
	Resource []Resources `json:"Resources"`
}

type SCIMIDResult struct {
	ScimID string `json:"scimid"`
}

type UserDetails struct {
	Username  string `json:"username"`
	KeySecret string `json:"keysecret"`
}

type Server struct {
	router *mux.Router
}
type Config struct {
	ISHost          string
	ISPort          string
	APIMHost        string
	APIMPort        string
	ListenServePort string
	ISUsername      string
	ISPassword      string
	UMHost          string
	UMPort          string
	KeySecret       string
}
