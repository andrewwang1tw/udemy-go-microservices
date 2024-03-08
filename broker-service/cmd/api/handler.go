package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"email"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

/*
type AuthPayload struct{
	Email 		string `json:"email"`
	Password	string `json:"password"`
}
*/

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// ////////////////////////////////////////////////////////////////////////////////////////////////
func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		//app.logItem(w, requestPayload.Log)			// log
		//app.logEventViaRabbit(w, requestPayload.Log) // log by MQ
		app.logEventViaRPC(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknow action"))
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////
func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	_ = app.writeJSON(w, http.StatusOK,
		jsonResponse{
			Error:   false,
			Message: "Hit the broker",
		},
	)
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, err := json.MarshalIndent(msg, "", "\t")
	if err != nil {
		log.Println("sendMail: Error marshalling data", err)
		app.errorJSON(w, err)
		
	}

	mailServerURL := "http://mail-service/send"
	request, err := http.NewRequest("POST", mailServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("sendMail: Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("sendMail: Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Println("sendMail error calling mail service", response.StatusCode)
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}

	// decode the json from the auth service
	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		log.Println("sendMail Error creating decoder", err)
		app.errorJSON(w, err)
		return
	}

	// send back json
	var payLoad jsonResponse
	payLoad.Error = false
	payLoad.Message = "Message send to " + msg.To

	app.writeJSON(w, http.StatusAccepted, payLoad)
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	logPayload := LogPayload{
		Name: l.Name,
		Data: l.Data,
	}

	j, _ := json.MarshalIndent(&logPayload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//
	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"
	app.writeJSON(w, http.StatusAccepted, payload)
}

/*
func (app *Config) pushToQueue(name string, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}
	payload := LogPayload{
		Name: name,
		Data: msg,
	}
	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}
*/

//////////////////////////////////////////////////////////////////////////////////////////////////
func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		log.Println("logItem: Error marshalling data", err)
	}

	serviceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("logItem: Error creating request", err)
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("logItem: Error getting response", err)
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusAccepted {
		log.Println("logItem: Status not Accepted ", response.StatusCode)
		app.errorJSON(w, errors.New("logItem: Status not Accepted"))
		return
	}

	var payLoad jsonResponse
	payLoad.Error = false
	payLoad.Message = "logged!"

	app.writeJSON(w, http.StatusAccepted, payLoad)
}

///////////////////////////////////////////////////////////////////////////////////////
type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logEventViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload {
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse {
		Error: false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
