package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"strconv"
	"time"
)

func main() {

	// Данные тестового письма
	mailData := EmailData{
		To:          os.Getenv("target"), // Кому отправляем (адрес получателя)
		Subject:     "Test",              // Тема письма
		ContentType: "text/plain",        // Тип контента (text/plain, text/html)
		Body:        "Test milky-mailer", // Тело письма
		FromId:      "milky-noreply",     // Идентификатор отправителя (согласно конфигурации мейлера)
	}

	cfg := NewConfig()

	// Установка соединения с AMQP
	connection, err := amqp.DialConfig(fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.AMQP.User,
		cfg.AMQP.Password,
		cfg.AMQP.Host,
		cfg.AMQP.Port),
		amqp.Config{
			Vhost: cfg.AMQP.VirtualHost,
		})
	if err != nil {
		panic(errors.Join(err, errors.New("error connect to amqp")))
	}

	// Отложенное закрытие соединения
	defer func(connection *amqp.Connection) {
		err := connection.Close()
		if err != nil {
			panic(errors.Join(err, errors.New("error close amqp connection")))
		}
	}(connection)

	// Создание канала
	amqpChannel, err := connection.Channel()
	if err != nil {
		panic(errors.Join(err, errors.New("error create amqp channel")))
	}

	// Проверка существования обменника. (Важно! В случае, если обменник не существует, то нельзя создавать новый обменник)
	err = amqpChannel.ExchangeDeclarePassive(
		cfg.AMQP.Exchange, // Имя обменника
		"direct",          // Тип обменника
		true,              // durable
		false,             // autoDelete
		false,             // internal
		false,             // noWait
		nil,               // args
	)
	if err != nil {
		panic(errors.Join(err, errors.New("error declare exchange")))
	}

	// Формирование сообщения
	amqpMessage := amqp.Publishing{
		Headers: amqp.Table{
			"To":      mailData.To,      // Кому отправляем (адрес получателя)
			"Subject": mailData.Subject, // Тема письма
			"FromId":  mailData.FromId,  // Идентификатор отправителя (согласно конфигурации мейлера)
		},
		ContentType: mailData.ContentType,  // Тип контента (text/plain, text/html)
		Body:        []byte(mailData.Body), // Тело письма

		// Необязательно, но очень настоятельно рекомендуется

		Priority: 2, // Приоритет сообщения

		// Время жизни сообщения (в данном примере 12 часов)
		Expiration: strconv.FormatInt((time.Hour * 12).Milliseconds(), 10),

		// Ниже идут необязательные поля, которые наполняют логи полезной информацией
		Timestamp: time.Now(),            // Время создания сообщения
		MessageId: uuid.New().String(),   // Идентификатор сообщения
		AppId:     "milky-mailer-client", // Идентификатор приложения отправителя (appName или другое)
	}

	// На этот контекст не смотрите, ничего важного для понимания работы с AMQP тут нет
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Публикация сообщения в очередь
	err = amqpChannel.PublishWithContext(
		ctx,
		cfg.AMQP.Exchange, // Имя обменника
		"",                // Имя очереди (пустая строка, т.к. очередь привязана к обменнику и её указание не требуется)
		false,             // mandatory
		false,             // immediate
		amqpMessage,       // Сообщение
	)
	if err != nil {
		panic(errors.Join(err, errors.New("error publish message")))
	}

	fmt.Println("Message published")

}
