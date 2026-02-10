package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func NewMySQL(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

func NewRedis(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}
	return &KafkaProducer{writer: w}, nil
}

func (p *KafkaProducer) Publish(ctx context.Context, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
