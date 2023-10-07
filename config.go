package main

type Config struct {
	AMQP AMQPConfig
}

type AMQPConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	Exchange    string
	VirtualHost string
}

type EmailData struct {
	To          string
	Subject     string
	ContentType string
	Body        string // TODO возможно, тут лучше передавать указатель. нужно проверить
	FromId      string
}

func NewConfig() *Config {
	return &Config{
		AMQP: AMQPConfig{
			Host:        "127.0.0.1",
			Port:        5672,
			User:        "mail",
			Password:    "mail",
			Exchange:    "milky-mailer-dev",
			VirtualHost: "/",
		},
	}
}
