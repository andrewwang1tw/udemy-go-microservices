package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {	
	var requestPayload struct {
		Email 		string	`json:"email"`
		Password 	string	`json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil{
		log.Println("Error readJSON", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	
	// validate the user against the dtabase 
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		log.Println("Error GetByEmail", requestPayload.Email)
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)		
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		log.Println("Error password", requestPayload.Password)
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return		
	}

	// log authentication
	if err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email)); err != nil{
		app.errorJSON(w, err)
		return	
	}

	payload := jsonResponse{
		Error: false,
		Message: fmt.Sprintf("Loggined in user %s", user.Email),
		Data: user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}


func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
        log.Println("authentication logItem: Error marshalling data", err)
    }

	serviceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("authentication logItem: Error creating request", err)
		return err
	}

	//request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		log.Println("authentication logItem: Error getting response", err)
		return err
	}

	return nil
}	