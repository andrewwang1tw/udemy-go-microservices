package event

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (e *Emitter) Push(event string, severity string) error {
	channel, err := e.connection.Channel()
	if err != nil{
		return err
	}
	defer channel.Close()
	
	// Create a context with a timeout (optional)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()	

	log.Println("Push to Channel")
	//err = channel.Publish("log_topics", severity, false, false, amqp.Publishing{ ContentType: "text/plain",Body: []byte(event),},)
	err = channel.PublishWithContext(ctx, "log_topics", severity, false, false, amqp.Publishing{ ContentType: "text/plain",Body: []byte(event),},)
	if err != nil {
		return err
	}

	return nil
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func NewEventEmitter(conn *amqp.Connection) (Emitter, error){
	emitter := Emitter{
		connection: conn,
	}

	/*
	if err := emitter.setup(); err != nil {
		return Emitter{}, err
	}
	*/

	channel, err := conn.Channel()
	if err != nil{
		return Emitter{}, err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare("log_topics",	"topic", true, false, false, false, nil)
	if err != nil {
		return Emitter{}, err
	}

	return emitter, nil
}

/*
func (e *Emitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil{
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare("log_topics",	"topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}
*/