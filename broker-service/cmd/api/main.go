package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "8080"

type Config struct{
	Rabbit *amqp.Connection
}

////////////////////////////////////////////////////////////////////////
func main() {	
////////////////////////////////////////////////////////////////////////
	// connect to rabbitmq
	rabbitConn, err := connectRabbitMQ()
	if err != nil {
		log.Println("rabbitmq connect error", err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	if err := srv.ListenAndServe(); err != nil{
		log.Panic(err)
	}

}



/////////////////////////////////////////////////////////////////////////////////////
func connectRabbitMQ() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1* time.Second
	var connection *amqp.Connection

	for {
		//c, err := amqp.Dial("amqp://guest:guest@localhost")
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("rabbitmq not yet ready... ", err)			
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = c
			break
		}
	
		if counts > 5 {
			log.Println("rabbitmq connect counts > 5 ", err)
			return nil, err
		}
	
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}
	
	return connection, nil
}