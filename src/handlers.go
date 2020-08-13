package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (s *Server) handleregisteruser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle Register User in IS with SCIM Has Been Called!")

		regUser := RegisterUser{}
		err := json.NewDecoder(r.Body).Decode(&regUser)

		//handle for bad JSON provided
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Improper registration details provided")
			return
		}

		if regUser.KeySecret != config.Key_Secret {
			keyErrorByte, _ := json.Marshal("Resource accessed without the correct key and secret!")
			w.WriteHeader(500)
			w.Write(keyErrorByte)
			fmt.Println("Resource accessed without the correct key and secret!")
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client := &http.Client{}
		var data = strings.NewReader(`{"schemas":[],"name":{"familyName":"` + regUser.Surname + `" ,"givenName":"` + regUser.Name + `"},"userName":"` + regUser.Username + `","password":"` + regUser.Password + `","emails":[{"primary":true,"value":"` + regUser.Email + `","type":"home"},{"value":"` + regUser.Email + `","type":"work"}]}`)

		req, err := http.NewRequest("POST", "https://"+config.IS_Host+":"+config.IS_Port+"/wso2/scim/Users", data)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "admin")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == 409 {
			userExists := RegisterUserResponse{}
			userExists.UserCreated = "false"
			userExists.Username = ""
			userExists.UserID = "00000000-0000-0000-0000-000000000000"
			userExists.Message = "This Username Already Exists!"

			js, jserr := json.Marshal(userExists)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to register user when that user already exists.")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var identityServerResponse IdentityServerResponse

		err = json.Unmarshal(bodyText, &identityServerResponse)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding registration response")
			return
		}

		requestByte, _ := json.Marshal(regUser)
		reqToUM, respErr := http.Post("http://"+config.UM_Host+":"+config.UM_Port+"/user", "application/json", bytes.NewBuffer(requestByte))

		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with User Manager service endpoint for request to register")
			return
		}
		if reqToUM.StatusCode != 200 {
			fmt.Fprint(w, "Request to DB can't be completed...")
			fmt.Println("Unable to process registration")
		}
		if reqToUM.StatusCode == 500 {
			w.WriteHeader(500)

			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintf(w, "Request to DB can't be completed..."+bodyString)
			fmt.Println("Request to DB can't be completed..." + bodyString)
			return
		}
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Registration is not able to be completed by internal error")
			return
		}

		//close the request
		defer reqToUM.Body.Close()

		var registerResponse RegisterUserResponse

		//decode request into decoder which converts to the struct
		decoder := json.NewDecoder(reqToUM.Body)

		err = decoder.Decode(&registerResponse)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding registration response")
			return
		}
		js, jserr := json.Marshal(registerResponse)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to register user")
			return
		}

		if registerResponse.UserCreated == "false" {
			client := &http.Client{}
			req, err := http.NewRequest("DELETE", "https://"+config.IS_Host+":"+config.IS_Port+"/wso2/scim/Users/"+identityServerResponse.ID, nil)
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Set("Accept", "application/json")
			req.SetBasicAuth("admin", "admin")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			bodyText, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", bodyText)
		}

		//return back to Front-End user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)

	}
}

func (s *Server) handleupdateuser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//get JSON payload
		updateUser := UpdateUser{}
		err := json.NewDecoder(r.Body).Decode(&updateUser)

		//handle for bad JSON provided
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println(err.Error())
			return
		}

		client := &http.Client{}

		//create byte array from JSON payload
		requestByte, _ := json.Marshal(updateUser)

		//put to crud service
		req, err := http.NewRequest("PUT", "http://"+config.UM_Host+":"+config.UM_Port+"/user", bytes.NewBuffer(requestByte))
		if err != nil {
			fmt.Fprint(w, err.Error())
			fmt.Println(err.Error())
			return
		}

		// Fetch Request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprint(w, err.Error())
			fmt.Println(err.Error())
			return
		}

		//close the request
		defer resp.Body.Close()

		//create new response struct
		var updateResponse UpdateUserResult

		decoder := json.NewDecoder(resp.Body)

		err = decoder.Decode(&updateResponse)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println(err.Error())
			return
		}

		//convert struct back to JSON
		js, jserr := json.Marshal(updateResponse)

		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println(err.Error())
			return
		}

		if updateResponse.UserUpdated == true {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			// TODO: Set InsecureSkipVerify as config in environment.env
			client1 := &http.Client{}
			var data = strings.NewReader(`{"schemas":[],"name":{"familyName":"` + updateUser.Surname + `" ,"givenName":"` + updateUser.Name + `"},"userName":"` + updateUser.Username + `","emails":[{"primary":true,"value":"` + updateUser.Email + `","type":"home"},{"value":"` + updateUser.Email + `","type":"work"}]}`)

			req2, err1 := http.NewRequest("PUT", "https://"+config.IS_Host+":"+config.IS_Port+"/wso2/scim/Users/"+updateUser.ScimID, data)
			if err != nil {
				log.Fatal(err)
			}
			req2.Header.Set("Content-Type", "application/json")
			req2.SetBasicAuth("admin", "admin")
			resp1, err1 := client1.Do(req2)
			if err1 != nil {
				log.Fatal(err1)
			}

			bodyText, err := ioutil.ReadAll(resp1.Body)
			if err != nil {
				log.Fatal(err)
			}

			defer resp1.Body.Close()

			var identityServerResponse IdentityServerResponse

			err = json.Unmarshal(bodyText, &identityServerResponse)

			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, err.Error())
				fmt.Println(err.Error())
				fmt.Println("Error occured in decoding registration response")
				return
			}

		}

		//return back to Front-End user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)

	}
}

func (s *Server) handleloginuser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle Login User in IS with SCIM Has Been Called!")

		//get JSON payload
		loginUser := LoginUser{}
		err := json.NewDecoder(r.Body).Decode(&loginUser)

		//handle for bad JSON provided
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println(err.Error())
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client := &http.Client{}
		data := url.Values{}
		data.Set("grant_type", "password")
		data.Add("username", loginUser.Username)
		data.Add("password", loginUser.Password)

		req, err := http.NewRequest("POST", "https://"+config.APIM_Host+":"+config.APIM_Port+"/token", bytes.NewBufferString(data.Encode()))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", "Basic "+config.Key_Secret)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == 400 {
			userExists := LoginUserResult{}
			userExists.UserLoggedIn = false
			userExists.Username = "None"
			userExists.Institution = "None"
			userExists.UserID = "00000000-0000-0000-0000-000000000000"
			userExists.Message = "Incorrect login credentials"

			js, jserr := json.Marshal(userExists)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to register user when incorrect login details were recieved.")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var identityServerResponse TokenResponse

		err = json.Unmarshal(bodyText, &identityServerResponse)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding registration response")
			return
		}

		reqtoUM, respErr := http.Get("http://" + config.UM_Host + ":" + config.UM_Port + "/userlogin?username=" + loginUser.Username + "&password=" + loginUser.Password)

		//check for response error of 500
		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with CRUD service endpoint for request to login user")
			return
		}
		if reqtoUM.StatusCode != 200 {
			fmt.Println("Request to DB can't be completed to login user")
		}
		if reqtoUM.StatusCode == 500 {
			w.WriteHeader(500)
			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintf(w, "Database error occured upon retrieval"+bodyString)
			fmt.Println("Database error occured upon retrieval" + bodyString)
			return
		}

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Logging in is not able to be completed by internal error")
			return
		}

		//close the request
		defer reqtoUM.Body.Close()
		var loginuseresult LoginUserResult

		//decode request into decoder which converts to the struct
		decoder := json.NewDecoder(reqtoUM.Body)

		loginuseresult.Accesstoken = identityServerResponse.Accesstoken
		loginuseresult.Refreshtoken = identityServerResponse.Refreshtoken

		err = decoder.Decode(&loginuseresult)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding registration response")
			return
		}
		js, jserr := json.Marshal(loginuseresult)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to register user")
			return
		}

		//return back to Front-End user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)

	}
}

func (s *Server) handlechangeuserpassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle Change password with SCIM Has Been Called!")

		updatePassword := UpdatePassword{}
		err := json.NewDecoder(r.Body).Decode(&updatePassword)

		//handle for bad JSON provided
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			return
		}

		client := &http.Client{}

		//create byte array from JSON payload
		requestByte, _ := json.Marshal(updatePassword)

		//put to crud service
		req, err := http.NewRequest("PUT", "http://"+config.UM_Host+":"+config.UM_Port+"/userpassword", bytes.NewBuffer(requestByte))
		if err != nil {
			fmt.Fprint(w, err.Error())
			fmt.Println("Error in communication with CRUD service endpoint for request to update user")
			return
		}

		// Fetch Request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		//close the request
		defer resp.Body.Close()

		//create new response struct
		var passwordResponse UpdatePasswordResult
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&passwordResponse)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			return
		}

		if passwordResponse.PasswordUpdated == false {

			js, jserr := json.Marshal(passwordResponse)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to register user when that user already exists.")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client1 := &http.Client{}
		var data = strings.NewReader(`{"schemas":[],"userName":"` + updatePassword.Username + `","password":"` + updatePassword.Password + `"}`)

		reqtoIS, err := http.NewRequest("PATCH", "https://"+config.IS_Host+":"+config.IS_Port+"/wso2/scim/Users/"+updatePassword.ScimID, data)
		if err != nil {
			log.Fatal(err)
		}
		reqtoIS.Header.Set("Content-Type", "application/json")
		reqtoIS.SetBasicAuth("admin", "admin")
		resp1, err := client1.Do(reqtoIS)

		if err != nil {
			log.Fatal(err)
		}

		if resp1.StatusCode == 500 {
			fmt.Println("Password revert to the old one has taken place.")
			//send new JSON payload
			revertPassword := UpdatePassword{}
			revertPassword.UserID = updatePassword.UserID
			revertPassword.ScimID = updatePassword.ScimID
			revertPassword.Username = updatePassword.Username
			revertPassword.CurrentPassword = updatePassword.Password
			revertPassword.Password = updatePassword.CurrentPassword
			err := json.NewDecoder(r.Body).Decode(&updatePassword)

			client := &http.Client{}

			//create byte array from JSON payload
			requestByte, _ := json.Marshal(revertPassword)

			//put to user manager
			req, err := http.NewRequest("PUT", "http://"+config.UM_Host+":"+config.UM_Port+"/userpassword", bytes.NewBuffer(requestByte))
			if err != nil {
				fmt.Fprint(w, err.Error())
				fmt.Println("Error in communication with CRUD service endpoint for request to update user")
				return
			}

			// Fetch Request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprint(w, err.Error())
				return
			}

			//close the request
			defer resp.Body.Close()

			var passwordrevert UpdatePasswordResult
			decoder := json.NewDecoder(resp.Body)

			passwordrevert.PasswordUpdated = false
			passwordrevert.Message = "An internal error has occured whilst trying to change our password. Please use your old password for now."

			err = decoder.Decode(&passwordResponse)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, err.Error())
				return
			}

			//convert struct back to JSON
			js, jserr := json.Marshal(passwordrevert)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to revert password for a user")
				return
			}

			//return back to Front-End user
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		bodyText, err1 := ioutil.ReadAll(resp1.Body)
		if err1 != nil {
			log.Fatal(err1)
		}

		var identityServerResponse IdentityServerResponse

		err1 = json.Unmarshal(bodyText, &identityServerResponse)

		if err1 != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding registration response")
			return
		}

		js, jserr := json.Marshal(passwordResponse)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to changing the user password.")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)
		return

	}

}

func (s *Server) handleforgotpassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle forgot password with SCIM Has Been Called!")
		//get user email from url
		email := r.URL.Query().Get("email")
		scimid := r.URL.Query().Get("scimid")
		//get userID from crud service
		req, respErr := http.Get("http://" + config.UM_Host + ":" + config.UM_Port + "/password?email=" + email)

		//check for response error of 500
		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with CRUD service endpoint for request to retrieve user information")
			return
		}
		if req.StatusCode != 200 {
			w.WriteHeader(500)
			fmt.Fprint(w, "Request to DB can't be completed...")
			fmt.Println("Request to DB can't be completed...")
		}
		if req.StatusCode == 500 {
			w.WriteHeader(500)
			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintf(w, "An internal error has occured whilst trying to get user data"+bodyString)
			fmt.Println("An internal error has occured whilst trying to get user data" + bodyString)
			return
		}

		defer req.Body.Close()

		//create new response struct
		var getResponse getPassword
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&getResponse)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("An internal error has occured whilst trying to decode the get user response")
			return
		}
		if getResponse.GotUser == false {
			var userResponse UserResult

			userResponse.Message = "A new password cannot be granted at this time as an appropriate email address has not been provided"

			js, jserr := json.Marshal(userResponse)
			if jserr != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, jserr.Error())
				fmt.Println("Error occured when trying to marshal the response to changing the user password.")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(js)
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		reqtoUM, err := http.Get("http://" + config.UM_Host + ":" + config.UM_Port + "/forgotpassword?email=" + email)
		if respErr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, respErr.Error())
			fmt.Println("Error in communication with CRUD service endpoint for request to retrieve password reset information")
			return
		}
		if reqtoUM.StatusCode != 200 {
			w.WriteHeader(reqtoUM.StatusCode)
			fmt.Fprint(w, "Request to DB can't be completed...")
			fmt.Println("Request to DB can't be completed...")
		}
		if reqtoUM.StatusCode == 500 {
			w.WriteHeader(500)
			bodyBytes, err := ioutil.ReadAll(reqtoUM.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Fprintf(w, "An internal error has occured whilst trying to get advertisement data"+bodyString)
			fmt.Println("An internal error has occured whilst trying to get advertisement data" + bodyString)
			return
		}

		//close the request
		defer reqtoUM.Body.Close()

		var passwordresetresult EmailResult

		//decode request into decoder which converts to the struct
		decoder1 := json.NewDecoder(reqtoUM.Body)
		err2 := decoder1.Decode(&passwordresetresult)
		if err2 != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding get Advertisement response ")
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client1 := &http.Client{}
		var data = strings.NewReader(`{"schemas":[],"userName":"` + getResponse.Username + `","password":"` + passwordresetresult.Password + `"}`)

		reqtoIS, err := http.NewRequest("PATCH", "https://"+config.IS_Host+":"+config.IS_Port+"/wso2/scim/Users/"+scimid, data)
		if err != nil {
			log.Fatal(err)
		}
		reqtoIS.Header.Set("Content-Type", "application/json")
		reqtoIS.SetBasicAuth("admin", "admin")
		resp1, err := client1.Do(reqtoIS)

		bodyText, err := ioutil.ReadAll(resp1.Body)
		if err != nil {
			log.Fatal(err)
		}

		if resp1.StatusCode == 500 {
			fmt.Println("Password revert to the old one has taken place.")
			//send new JSON payload
			revertPassword := UpdatePassword{}
			revertPassword.UserID = getResponse.UserID
			revertPassword.ScimID = scimid
			revertPassword.Username = getResponse.Username
			revertPassword.CurrentPassword = passwordresetresult.Password
			revertPassword.Password = getResponse.Password

			client := &http.Client{}

			//create byte array from JSON payload
			requestByte, _ := json.Marshal(revertPassword)

			//put to user manager
			req, err := http.NewRequest("PUT", "http://"+config.UM_Host+":"+config.UM_Port+"/userpassword", bytes.NewBuffer(requestByte))
			if err != nil {
				fmt.Fprint(w, err.Error())
				fmt.Println("Error in communication with CRUD service endpoint for request to update user")
				return
			}

			// Fetch Request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprint(w, err.Error())
				return
			}

			//close the request
			defer resp.Body.Close()

			var passwordResponse UpdatePasswordResult
			decoder := json.NewDecoder(req.Body)
			err = decoder.Decode(&passwordResponse)

			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, err.Error())
				return
			}
			if passwordResponse.PasswordUpdated == true {
				var passwordrevert UserResult
				passwordrevert.Message = "An internal error has occured whilst trying to generate a new password. Please use your old password for now."
				//convert struct back to JSON
				js, jserr := json.Marshal(passwordrevert)
				if jserr != nil {
					w.WriteHeader(500)
					fmt.Fprint(w, jserr.Error())
					fmt.Println("Error occured when trying to marshal the response to revert password for a user")
					return
				}

				//return back to Front-End user
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write(js)
				return

			}

		}

		var identityServerResponse IdentityServerResponse

		err = json.Unmarshal(bodyText, &identityServerResponse)

		var userResponse UserResult

		userResponse.Message = passwordresetresult.Message

		js, jserr := json.Marshal(userResponse)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the response to changing the user password.")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)
		return

	}

}

func (s *Server) handlegetscimid() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle get SCIM ID Has Been Called!")
		userDetails := UserDetails{}
		err := json.NewDecoder(r.Body).Decode(&userDetails)

		//handle for bad JSON provided.
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Improper registration details provided")
			return
		}

		if userDetails.KeySecret != config.Key_Secret {
			keyErrorByte, _ := json.Marshal("Resource accessed without the correct key and secret!")
			w.WriteHeader(500)
			w.Write(keyErrorByte)
			fmt.Println("Resource accessed without the correct key and secret!")
			return
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client1 := &http.Client{}

		reqtoIS, err := http.NewRequest("GET", "https://"+config.IS_Host+":"+config.IS_Port+"/wso2/scim/Users?filter=userName+Eq+%22"+userDetails.Username+"%22", nil)
		if err != nil {
			log.Fatal(err)
		}
		reqtoIS.SetBasicAuth("admin", "admin")
		resp1, err := client1.Do(reqtoIS)

		bodyText, err := ioutil.ReadAll(resp1.Body)
		if err != nil {
			log.Fatal(err)
		}

		var identityServerSCIMID IdentityServerSCIMID

		err = json.Unmarshal(bodyText, &identityServerSCIMID)

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
			fmt.Println("Error occured in decoding SCIM ID response")
			return
		}

		js, jserr := json.Marshal(identityServerSCIMID)
		if jserr != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, jserr.Error())
			fmt.Println("Error occured when trying to marshal the user SCIM ID")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(js)
		return

	}

}

/* ========================================================================
===========================================================================
=========================================================================*/
