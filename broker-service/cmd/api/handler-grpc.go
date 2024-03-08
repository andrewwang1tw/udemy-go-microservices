package main

import (

	"broker/logs"
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {

	var requestPayload RequestPayload

	if err := app.readJSON(w, r, &requestPayload); err != nil {
		app.errorJSON(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	grpcClient := logs.NewLogServiceClient(conn)
	_, err = grpcClient.WriteLog(ctx, &logs.LogRequest{
			LogEntry: &logs.Log{
				Name: requestPayload.Log.Name,
				Data: requestPayload.Log.Data,
			},
		},
	)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"
	app.writeJSON(w, http.StatusAccepted, payload)
}
