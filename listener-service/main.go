package main

import (
	//"fmt"
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println("rabbitmq connect error", err)
		os.Exit(1)
	}
	defer rabbitConn.Close()	
	
	/* create consumer
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		log.Println("event.NewConsumer(rabbitConn): ", err)
		panic(err)
	}
	*/
	
	// watch the queue and consume events
	err = event.Listen(rabbitConn, []string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println("event Listening: ", err)
	}	
}


/////////////////////////////////////////////////////////////////////////////////////
func connect() (*amqp.Connection, error) {
/////////////////////////////////////////////////////////////////////////////////////	
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