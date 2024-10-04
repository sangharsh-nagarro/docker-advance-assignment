package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	redisClient *redis.Client
	mongoClient *mongo.Client
)

type User struct {
	Name  string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
}

func initRedis() {
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisURL := fmt.Sprintf("redis://:%s@redis:6379/0", redisPassword)

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("Failed to parse Redis URL", "error", err)
		os.Exit(1)
	}

	redisClient = redis.NewClient(opts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Redis successfully")
}

func initMongo() {
	username := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	password := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
	mongoURL := fmt.Sprintf("mongodb://%s:%s@mongo:27017/development?authSource=admin", username, password)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		slog.Error("Failed to ping MongoDB", "error", err)
		os.Exit(1)
	}

	mongoClient = client
	slog.Info("Connected to MongoDB successfully")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collection := mongoClient.Database("development").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		slog.Error("Failed to insert user", "error", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"id":      result.InsertedID,
	})
	if err != nil {
		panic(err)
	}
}

type UserResponse struct {
	User       User
	DataSource string
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	var dataSource string

	// Try to get user from Redis cache
	cachedUser, err := redisClient.Get(ctx, "user:"+email).Result()
	if err == nil {
		// User found in cache
		err = json.Unmarshal([]byte(cachedUser), &user)
		if err == nil {
			dataSource = "cache"
		}
	}

	if dataSource == "" {
		// User not found in cache, query MongoDB
		collection := mongoClient.Database("development").Collection("users")
		err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				slog.Error("Failed to fetch user", "error", err)
				http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
			}
			return
		}
		dataSource = "database"

		// Cache the user in Redis
		userJSON, _ := json.Marshal(user)
		err = redisClient.Set(ctx, "user:"+email, userJSON, 1*time.Hour).Err()
		if err != nil {
			slog.Error("Failed to cache user", "error", err)
		}
	}

	response := UserResponse{
		User:       user,
		DataSource: dataSource,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		panic(err)
	}
}

func main() {
	initRedis()
	initMongo()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("<h1>healthy</h1>"))
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("POST /api/user/create", createUserHandler)
	mux.HandleFunc("GET /api/user", getUserHandler)

	slog.Info("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
