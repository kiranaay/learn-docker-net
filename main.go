package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	redisClient *redis.Client
	mongoClient *mongo.Client
)

func initRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPort := os.Getenv("REDIS_PORT")
	if redisAddr == "" {
		redisAddr = "localhost"
	}
	if redisPort == "" {
		redisPort = "6379"
	}
	redisAddr = redisAddr + ":" + redisPort
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", //no pasword set
		DB:       0,  //use default DB
	})
}

func initMongo() {
	mongoAddr := os.Getenv("MONGO_ADDR")
	mongoPort := os.Getenv("MONGO_PORT")
	if mongoAddr == "" {
		mongoAddr = "localhost"
	}
	if mongoPort == "" {
		mongoPort = "27017"
	}
	mongoAddr = "mongodb://" + mongoAddr + ":" + mongoPort
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoAddr))
	if err != nil {
		fmt.Printf("Failed to create MongoDB client: %v\n", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Failed to create MongoDB client: %v\n", err)
		return
	}
	mongoClient = client
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	message := os.Getenv("HELL0_MESSAGE")
	if message == "" {
		message = "Hello, World! (no set env)"
	}
	fmt.Fprintf(w, message)
}

func checkRedisConnection() error {
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	return err
}

func checkMongoDBConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mongoClient.Ping(ctx, nil)
	return err
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	var status string
	if err := checkRedisConnection(); err != nil {
		status = "Redis connection failed"
	} else if err := checkMongoDBConnection(); err != nil {
		status = "MongoDB connection failed"
	} else {
		status = "OK"
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, status)
}

func main() {
	initRedis()
	initMongo()

	http.HandleFunc("/", helloWorld)
	http.HandleFunc("/health", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Server running at http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
