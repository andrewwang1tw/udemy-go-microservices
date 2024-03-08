package main

import (	
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type AuthPayload struct{
	Email 		string `json:"email"`
	Password	string `json:"password"`
}



////////////////////////////////////////////////////////////////////////////////////
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload){
	// Create some json we'll send to auth microservice
	jsonData, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
        log.Println("authenticate: Error marshalling data", err)
    }

	// Call the microservice	
	const serviceURL = "http://authentication-service/authenticate"
	request, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("authenticate: Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code	
	if response.StatusCode == http.StatusUnauthorized {
		log.Println("Wrong status code", response.StatusCode)
		app.errorJSON(w, errors.New("invalid credential"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		log.Println("Wrong status code 2", response.StatusCode)
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}	

	// decode the json from the auth service
	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		log.Println("Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Println("Error from jsonFromService", jsonFromService.Error)
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	// send back json
	var payLoad jsonResponse
	payLoad.Error = false
	payLoad.Message = "Authenticated!"
	payLoad.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payLoad)
}