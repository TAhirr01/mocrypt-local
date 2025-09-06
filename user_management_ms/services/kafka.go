package services

import (
	"encoding/json"
	"log"
	"user_management_ms/dtos/request"

	"github.com/IBM/sarama"
)

func SendVerifyPhoneNumberEventToKafka(verifyPhoneEvent *request.VerifyPhoneEvent) error {
	phoneData, err := json.Marshal(verifyPhoneEvent)
	if err != nil {
		return err
	}
	producer, err := sarama.NewSyncProducer([]string{"localhost:9093"}, nil)
	if err != nil {
		log.Println("Failed to create sync producer:", err)
		return err
	}
	defer producer.Close()
	phoneMsg := &sarama.ProducerMessage{
		Topic: "VerifyPhoneEvent",
		Value: sarama.StringEncoder(phoneData),
	}
	partition, offset, err := producer.SendMessage(phoneMsg)
	if err != nil {
		log.Println("Failed to send phone:", err)
		return err
	}
	log.Printf("Successfully sent phone to partition %d at offset %d\n", partition, offset)
	return nil
}

func SendVerifyEmailEventToKafka(verifyEmailEvent *request.VerifyEmailEvent) error {
	emailData, err := json.Marshal(verifyEmailEvent)
	if err != nil {
		return err
	}

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Println("Failed to create sync producer:", err)
		return err
	}
	defer producer.Close()

	emailMsg := &sarama.ProducerMessage{
		Topic: "VerifyEmailEvent",
		Value: sarama.StringEncoder(emailData),
	}
	partition, offset, err := producer.SendMessage(emailMsg)
	if err != nil {
		log.Println("Failed to send email:", err)
	}
	log.Printf("Successfully sent email to partition %d at offset %d\n", partition, offset)
	return nil
}

func SendVerifyEmailAndPhoneNumberEvent(verifyEmailEvent *request.VerifyEmailEvent, verifyPhoneEvent *request.VerifyPhoneEvent) error {
	if err := SendVerifyPhoneNumberEventToKafka(verifyPhoneEvent); err != nil {
		return err
	}
	if err := SendVerifyEmailEventToKafka(verifyEmailEvent); err != nil {
		return err
	}
	return nil
}
