package event

import (
	"bytes"
	"encoding/json"
	"errors"
	//"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *amqp.Connection
	//queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}



///////////////////////////////////////////////////////////////////////////////////////
func Listen(conn *amqp.Connection, topics []string) error {
///////////////////////////////////////////////////////////////////////////////////////
	consumer := &Consumer{
		conn: conn,
	}

	channel, err := consumer.conn.Channel()
	if err != nil{
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare("log_topics",	"topic", true, false, false, false,	nil)
	if err != nil {
		return err
	}

	q, err := channel.QueueDeclare("", false, false,	true, false, nil)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err := channel.QueueBind(q.Name,	s,	"log_topics", false, nil)		
		if err != nil{
			return err
		}
	}

	messages, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func(){
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)
			
			go handlePayload(payload)
		}
	}()

	log.Printf("Waiting for message [Exchange, Queue], [logs_topic, %s]\n", q.Name)	
	<-forever

	return nil
}

////////////////////////////////////////////////////////////////////////////////////
func handlePayload (payload Payload) {
////////////////////////////////////////////////////////////////////////////////////	
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auth":
		// authenticate

	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////////////
func logEvent(entry Payload) error {
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
        log.Println("logItem: Error marshalling data", err)
    }

	serviceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("logItem: Error creating request", err)
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("logItem: Error getting response", err)
		return err
	}
	defer response.Body.Close()

	// make sure we get back the correct status code	
	if response.StatusCode != http.StatusAccepted {
		log.Println("logItem: Status not Accepted ", response.StatusCode)
		return errors.New("logItem: Status not Accepted")
	}
	
	return nil
}


/*
func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil{
		return err
	}
	return declareExchange(channel)
}
*/

/*
//////////////////////////////////////////////////////////////////////////////////
func NewConsumer(conn *amqp.Connection) (Consumer, error) {
//////////////////////////////////////////////////////////////////////////////////	
	consumer := Consumer{
		conn: conn,
	}

	channel, err := consumer.conn.Channel()
	if err != nil{
		return Consumer{}, err
	}

	err = channel.ExchangeDeclare(
		"log_topics",	// name
		"topic",		// type
		true,			// durable,
		false,			// auto-deleted ?
		false,			// internal  ?
		false,			// no-wait ?
		nil,			// arguments ?
	)
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}
*/