package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

		if regUser.KeySecret != config.KeySecret {
			keyErrorByte, _ := json.Marshal("Resource accessed without the correct key and secret!")
			w.WriteHeader(500)
			w.Write(keyErrorByte)
			fmt.Println("Resource accessed without the correct key and secret!")
			return
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		// TODO: Set InsecureSkipVerify as config in environment.env
		client := &http.Client{}
		var data = strings.NewReader(`{"schemas":[],"name":{"familyName":"` + regUser.Surname + `" ,"givenName":"` + regUser.Name + `"},"userName":"` + regUser.Username + `","password":"` + regUser.Password + `","emails":[{"primary":true,"value":"` + regUser.Email + `","type":"home"},{"value":"` + regUser.Email + `","type":"work"}]}`)

		req, err := http.NewRequest("POST", "https://auth.studymoney.co.za:9445/wso2/scim/Users", data)
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
		reqToUM, respErr := http.Post("http://"+config.UMHost+":"+config.UMPort+"/user", "application/json", bytes.NewBuffer(requestByte))

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
			req, err := http.NewRequest("DELETE", "https://auth.studymoney.co.za:9445/wso2/scim/Users/"+identityServerResponse.ID, nil)
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

/* ========================================================================
===========================================================================
=========================================================================*/
