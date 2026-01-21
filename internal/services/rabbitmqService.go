package services

import (
	"fmt"
	"log"

	"github.com/unbot2313/go-streaming-service/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQService define la interfaz para comunicarse con RabbitMQ
type RabbitMQService interface {
	Connect() error
	Close()
	Publish(queueName string, message []byte) error
	Consume(queueName string) error
}

// RabbitMQServiceImp es la implementación del servicio
type RabbitMQServiceImp struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewRabbitMQService crea una nueva instancia del servicio
func NewRabbitMQService() RabbitMQService {
	return &RabbitMQServiceImp{}
}

// failOnError maneja los errores de RabbitMQ usando log
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// logError registra errores sin detener la ejecución
func logError(err error, msg string) bool {
	if err != nil {
		log.Printf("ERROR - %s: %s", msg, err)
		return true
	}
	return false
}

// Connect establece la conexión con RabbitMQ
func (r *RabbitMQServiceImp) Connect() error {
	cfg := config.GetConfig()

	// Construir la URL de conexión: amqp://user:password@host:port/
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQUser,
		cfg.RabbitMQPassword,
		cfg.RabbitMQHost,
		cfg.RabbitMQPort,
	)

	// Establecer conexión
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("error al conectar con RabbitMQ: %w", err)
	}
	r.connection = conn

	// Crear canal
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error al crear canal: %w", err)
	}
	r.channel = ch

	log.Println("Conectado a RabbitMQ exitosamente")
	return nil
}

// Close cierra la conexión y el canal
func (r *RabbitMQServiceImp) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}
	log.Println("Conexión a RabbitMQ cerrada")
}

// Publish envía un mensaje a una cola
func (r *RabbitMQServiceImp) Publish(queueName string, message []byte) error {
	// Declarar la cola (se crea si no existe)
	queue, err := r.channel.QueueDeclare(
		queueName, // nombre
		false,     // durable: la cola NO sobrevive al reinicio del servidor
		false,     // autoDelete: NO se elimina cuando no hay consumidores
		false,     // exclusive: NO es exclusiva de esta conexión
		false,     // noWait: esperar confirmación del servidor
		nil,       // arguments: sin argumentos extra
	)
	if logError(err, "Error al declarar la cola") {
		return err
	}

	// Publicar mensaje
	err = r.channel.Publish(
		"",         // exchange: usamos el exchange por defecto
		queue.Name, // routing key: nombre de la cola
		false,      // mandatory: NO requerir que exista una cola
		false,      // immediate: NO requerir un consumidor activo
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
	if logError(err, "Error al publicar mensaje") {
		return err
	}

	log.Printf("Mensaje enviado a cola '%s': %s", queueName, string(message))
	return nil
}

// Consume escucha mensajes de una cola y los loguea
func (r *RabbitMQServiceImp) Consume(queueName string) error {
	// Declarar la cola
	queue, err := r.channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if logError(err, "Error al declarar la cola") {
		return err
	}

	// Registrar consumidor
	messages, err := r.channel.Consume(
		queue.Name, // cola
		"",         // consumer: nombre vacío = generado automáticamente
		true,       // autoAck: confirmar automáticamente los mensajes
		false,      // exclusive: NO exclusivo
		false,      // noLocal: permitir mensajes del mismo conexión
		false,      // noWait: esperar confirmación
		nil,        // arguments
	)
	if logError(err, "Error al registrar consumidor") {
		return err
	}

	log.Printf("Esperando mensajes en cola '%s'...", queueName)

	// Escuchar mensajes en un goroutine
	go func() {
		for msg := range messages {
			log.Printf("Mensaje recibido de '%s': %s", queueName, string(msg.Body))
		}
	}()

	return nil
}
