package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"notification-ms/config"
	"notification-ms/dtos"

	"github.com/IBM/sarama"
)

type EmailService struct {
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (es *EmailService) SendVerifyUserEmail(req *dtos.VerifyEmailEvent) error {
	from := config.Conf.Application.Smtp.From
	password := config.Conf.Application.Smtp.Password

	smtpHost := config.Conf.Application.Smtp.Host
	smtpPort := config.Conf.Application.Smtp.Port
	log.Println("SMTP Port:", smtpPort)

	subject := "Subject: Email Verification\n"
	log.Println(req.Otp)

	body := fmt.Sprintf("Salam,\n\nEmailini otp-ni appda yaz:%s\n", req.Otp)
	message := []byte(subject + "\n" + body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{req.Email}, message)
	if err != nil {
		log.Println("Email göndərmə xətası:", err)
		return err
	}
	log.Println("Email göndərildi:", req.Email)
	return nil
}

func (es *EmailService) ConsumeVerifyUserEvents() {
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatal("Failed to start Kafka consumer:", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition("VerifyEmailEvent", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal("Failed to consume partition:", err)
	}
	defer partitionConsumer.Close()

	for msg := range partitionConsumer.Messages() {
		var event *dtos.VerifyEmailEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("Invalid message:", err)
			continue
		}
		if err := es.SendVerifyUserEmail(&dtos.VerifyEmailEvent{Email: event.Email, Otp: event.Otp}); err != nil {
			log.Println(err)
			return
		}
		log.Printf("Send email to %s", event.Email)
	}
}
