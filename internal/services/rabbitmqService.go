package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/unbot2313/go-streaming-service/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// MaxRetries es el número máximo de reintentos antes de descartar el mensaje
	MaxRetries = 3
	// RetryDelay es el tiempo de espera entre reintentos
	RetryDelay = 5 * time.Second
)

// MessageHandler es una función que procesa un mensaje recibido
// Retorna error si el procesamiento falla (el mensaje será reenviado)
type MessageHandler func(message []byte) error

// RabbitMQService define la interfaz para comunicarse con RabbitMQ
type RabbitMQService interface {
	Connect() error
	Close()
	Publish(queueName string, message []byte) error
	Consume(queueName string, handler MessageHandler) error
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

// logError registra errores sin detener la ejecución
func logError(err error, msg string) bool {
	if err != nil {
		log.Printf("ERROR - %s: %s", msg, err)
		return true
	}
	return false
}

// getRetryCount extrae el contador de reintentos de los headers del mensaje
func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if count, ok := headers["x-retry-count"].(int32); ok {
		return int(count)
	}
	return 0
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

// Publish envía un mensaje a una cola con persistencia
func (r *RabbitMQServiceImp) Publish(queueName string, message []byte) error {
	// Declarar la cola durable (sobrevive reinicios de RabbitMQ)
	queue, err := r.channel.QueueDeclare(
		queueName, // nombre
		true,      // durable: la cola sobrevive al reinicio del servidor
		false,     // autoDelete: NO se elimina cuando no hay consumidores
		false,     // exclusive: NO es exclusiva de esta conexión
		false,     // noWait: esperar confirmación del servidor
		nil,       // arguments: sin argumentos extra
	)
	if logError(err, "Error al declarar la cola") {
		return err
	}

	// Context con timeout de 5 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publicar mensaje con persistencia
	err = r.channel.PublishWithContext(
		ctx,
		"",         // exchange: usamos el exchange por defecto
		queue.Name, // routing key: nombre de la cola
		false,      // mandatory: NO requerir que exista una cola
		false,      // immediate: NO requerir un consumidor activo
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Mensaje persistente (guardado en disco)
			ContentType:  "text/plain",
			Body:         message,
		},
	)
	if logError(err, "Error al publicar mensaje") {
		return err
	}

	log.Printf("Mensaje enviado a cola '%s': %s", queueName, string(message))
	return nil
}

// Consume escucha mensajes de una cola y los procesa con el handler proporcionado
// El handler debe retornar nil si el procesamiento fue exitoso, o error si falló
// Si el handler falla, el mensaje será reenviado a otro worker (Nack)
func (r *RabbitMQServiceImp) Consume(queueName string, handler MessageHandler) error {
	// Declarar la cola durable (debe coincidir con el publisher)
	queue, err := r.channel.QueueDeclare(
		queueName,
		true,  // durable: la cola sobrevive al reinicio del servidor
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if logError(err, "Error al declarar la cola") {
		return err
	}

	// Configurar QoS: solo recibir 1 mensaje a la vez
	// Esto distribuye el trabajo equitativamente entre workers
	err = r.channel.Qos(
		1,     // prefetch count: procesar 1 mensaje a la vez
		0,     // prefetch size: sin límite de bytes
		false, // global: aplica solo a este consumer
	)
	if logError(err, "Error al configurar QoS") {
		return err
	}

	// Registrar consumidor con acknowledgment manual
	messages, err := r.channel.Consume(
		queue.Name, // cola
		"",         // consumer: nombre vacío = generado automáticamente
		false,      // autoAck: FALSE - confirmaremos manualmente después de procesar
		false,      // exclusive: NO exclusivo (permite múltiples workers)
		false,      // noLocal: permitir mensajes del mismo conexión
		false,      // noWait: esperar confirmación
		nil,        // arguments
	)
	if logError(err, "Error al registrar consumidor") {
		return err
	}

	log.Printf("Worker esperando mensajes en cola '%s'...", queueName)

	// Escuchar mensajes en un goroutine
	go func() {
		for msg := range messages {
			retryCount := getRetryCount(msg.Headers)
			log.Printf("Mensaje recibido de '%s' (intento %d/%d)", queueName, retryCount+1, MaxRetries)

			// Procesar el mensaje con el handler
			err := handler(msg.Body)

			if err != nil {
				if retryCount >= MaxRetries-1 {
					// Máximo de reintentos alcanzado, descartar mensaje
					log.Printf("Máximo de reintentos alcanzado (%d). Descartando mensaje.", MaxRetries)
					msg.Ack(false)
				} else {
					// Reintentar: Ack el mensaje actual y republicar con contador incrementado
					log.Printf("Error procesando mensaje: %s - Reintentando en %v... (%d/%d)",
						err, RetryDelay, retryCount+2, MaxRetries)
					msg.Ack(false)

					// Esperar antes de reintentar
					time.Sleep(RetryDelay)

					// Republicar con retry count incrementado
					r.republishWithRetry(queueName, msg.Body, retryCount+1)
				}
			} else {
				log.Printf("Mensaje procesado exitosamente")
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// republishWithRetry republica un mensaje con el contador de reintentos incrementado
func (r *RabbitMQServiceImp) republishWithRetry(queueName string, body []byte, retryCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Headers: amqp.Table{
				"x-retry-count": int32(retryCount),
			},
			Body: body,
		},
	)
}
