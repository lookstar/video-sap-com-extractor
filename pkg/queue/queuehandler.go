package queue

import (
	"github.com/streadway/amqp"
	"log"
	"fmt"
	"github.com/lookstar/video-sap-com-extractor/pkg/collector"
	"io/ioutil"
	"encoding/json"
)

// it's a queue wrapper
type QueueHandler struct {
}

type MQCredential struct {
	URL string	`json:"url"`
}

func NewQueueHandler() *QueueHandler {
	return &QueueHandler{
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (p *QueueHandler) ReadCredential() *MQCredential {
	content, err := ioutil.ReadFile("./data/rabbitmq.json")
	if err != nil {
		fmt.Println("ReadCredential " + err.Error())
		panic(err)
	}
	ret := &MQCredential{}
	json.Unmarshal(content, ret)
	return ret
}

func (p *QueueHandler) Run() {
	mqCredential := p.ReadCredential()
	conn, err := amqp.Dial(mqCredential.URL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"video_queue", 
		true, 
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,
		0,
		false,
	)

	failOnError(err, "Failed to set Qos")

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil,)

	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func(){
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			videoURL := string(d.Body)
			fmt.Println(videoURL)
			dataCollector := collector.NewCollectorProvider(videoURL)
			err := dataCollector.DoWork()
			if err != nil {
				log.Printf("Error in %v %v", videoURL, err)
				d.Nack(false, true)
				log.Printf("Failed")
				return
			} else {
				log.Printf("Done")
				d.Ack(false)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
