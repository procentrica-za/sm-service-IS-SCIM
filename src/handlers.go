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
		fmt.Printf("%s\n", bodyText)

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

		username := r.URL.Query().Get("username")
		password := r.URL.Query().Get("password")
		if username == "" {
			w.WriteHeader(500)
			fmt.Fprint(w, "No username provided in URL")
			fmt.Println("A username has not been provided in URL")
			return
		}
		if password == "" {
			w.WriteHeader(500)
			fmt.Fprint(w, "No password provided in URL")
			fmt.Println("A password has not been provided in URL")
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client := &http.Client{}
		data := url.Values{}
		data.Set("grant_type", "password")
		data.Add("username", username)
		data.Add("password", password)

		req, err := http.NewRequest("POST", "https://"+config.APIM_Host+":"+config.APIM_Port+"/token", bytes.NewBufferString(data.Encode()))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", "Basic UTBPWkc1U0VmdVFFQWxGeFRheDg4bEEycWVrYTpqdVhlUDRKWnJJX0ZXOGxseUFpX2ZudFhDVjBh")
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

		reqtoUM, respErr := http.Get("http://" + config.UM_Host + ":" + config.UM_Port + "/userlogin?username=" + username + "&password=" + password)

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

/* ========================================================================
===========================================================================
=========================================================================*/
