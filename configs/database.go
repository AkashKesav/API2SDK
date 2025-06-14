package configs

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

// InitDatabase initializes the MongoDB connection using the provided configuration.
// This function replaces the previous ConnectMongoDB and uses the Config struct.
func InitDatabase(cfg *Config) error {
	log.Printf("Attempting to connect to MongoDB at %s, database: %s", cfg.MongoDBURI, cfg.MongoDBName)

	clientOptions := options.Client().
		ApplyURI(cfg.MongoDBURI).
		SetServerSelectionTimeout(60 * time.Second).
		SetConnectTimeout(30 * time.Second).
		SetSocketTimeout(120 * time.Second).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(300 * time.Second).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetHeartbeatInterval(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var client *mongo.Client
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			break
		}
		log.Printf("MongoDB connection attempt %d failed: %v", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * 2 * time.Second) // Exponential backoff
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB after %d retries: %w", maxRetries, err)
	}

	// Ping the database to verify connection
	for i := 0; i < maxRetries; i++ {
		err = client.Ping(ctx, nil)
		if err == nil {
			break
		}
		log.Printf("MongoDB ping attempt %d failed: %v", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to ping MongoDB after %d retries: %w", maxRetries, err)
	}

	DB = client.Database(cfg.MongoDBName)
	log.Println("Successfully connected to MongoDB and database selected.")

	// Test database operations (optional, can be removed or made conditional)
	if err := testDatabaseOperations(); err != nil {
		log.Printf("Warning: Database operations test failed: %v", err)
		// Depending on strictness, you might want to return an error here
		// return fmt.Errorf("database operations test failed: %w", err)
	} else {
		log.Println("Database operations test passed successfully.")
	}

	return nil
}

// GetDatabase returns the database instance
func GetDatabase() *mongo.Database {
	return DB
}

// GetCollection returns a collection instance
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}

// testDatabaseOperations performs basic operations to test database connectivity
func testDatabaseOperations() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test listing collections with proper filter
	filter := make(map[string]interface{})
	_, err := DB.ListCollectionNames(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	return nil
}
