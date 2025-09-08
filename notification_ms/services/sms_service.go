package services

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/kevinburke/twilio-go"
	"log"
	"notification-ms/config"
	"notification-ms/dtos"
)

type SmsService struct{}

func NewSmsService() *SmsService {
	return &SmsService{}
}

var client = twilio.NewClient(config.Conf.Application.Twilio.AccountSid, config.Conf.Application.Twilio.AuthToken, nil)

func (service *SmsService) SendVerifyPhoneEvent(req *dtos.VerifyPhoneEvent) error {
	_, err := client.Messages.SendMessage(config.Conf.Application.Twilio.From, req.Phone, req.Otp, nil)
	if err != nil {
		log.Println("Twilio error:", err)
		return err
	}
	return nil
}

func (service *SmsService) ConsumeVerifyPhoneEvents() {
	consumer, err := sarama.NewConsumer([]string{"localhost:9093"}, nil)
	if err != nil {
		log.Fatal("Failed to start Kafka consumer:", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("VerifyPhoneEvent", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal("Failed to consume partition:", err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		var event *dtos.VerifyPhoneEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("Invalid message:", err)
			continue
		}
		if err := service.SendVerifyPhoneEvent(&dtos.VerifyPhoneEvent{Phone: event.Phone, Otp: event.Otp}); err != nil {
			log.Println(err)
			return
		}
		log.Printf("Send email to %s", event.Phone)
	}
}
