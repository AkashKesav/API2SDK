package configs

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database
var MongoClient *mongo.Client

// InitDatabase initializes the MongoDB connection using the provided configuration.
// This function replaces the previous ConnectMongoDB and uses the Config struct.
func InitDatabase(cfg *Config) error {
	log.Printf("Attempting to connect to MongoDB at %s, database: %s", cfg.MongoDBURI, cfg.MongoDBName)

	// Simplified command monitoring - only log failures
	monitor := &event.CommandMonitor{
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			log.Printf("MongoDB command failed: %s (duration: %v, error: %v)", evt.CommandName, evt.Duration, evt.Failure)
		},
	}

	// Simplified pool monitoring - only log important events
	poolMonitor := &event.PoolMonitor{
		Event: func(evt *event.PoolEvent) {
			switch evt.Type {
			case event.PoolCreated:
				log.Printf("MongoDB connection pool created for %s", evt.Address)
			case event.PoolCleared:
				log.Printf("MongoDB connection pool cleared for %s", evt.Address)
			}
		},
	}

	// Simplified and more reliable client options
	clientOptions := options.Client().
		ApplyURI(cfg.MongoDBURI).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetSocketTimeout(30 * time.Second).
		SetMaxPoolSize(10).                   // Reduced from 50
		SetMinPoolSize(1).                    // Reduced from 5
		SetMaxConnIdleTime(30 * time.Second). // Reduced from 180
		SetRetryWrites(true).
		SetRetryReads(true).
		SetHeartbeatInterval(10 * time.Second).
		SetMonitor(monitor).
		SetPoolMonitor(poolMonitor)
		// Remove explicit TLS config - let MongoDB driver handle it automatically

	// Single connection attempt with proper context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping with a fresh context
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()

	log.Println("Pinging MongoDB...")
	err = client.Ping(pingCtx, nil)
	if err != nil {
		// Clean up the client if ping fails
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		client.Disconnect(disconnectCtx)
		disconnectCancel()
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	MongoClient = client
	DB = client.Database(cfg.MongoDBName)
	log.Println("Successfully connected to MongoDB and database selected.")

	// Test database operations with a fresh context
	testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer testCancel()

	if err := testDatabaseOperations(testCtx); err != nil {
		log.Printf("Warning: Database operations test failed: %v", err)
		// Don't fail initialization for test failures - just warn
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
func testDatabaseOperations(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Test listing collections with proper filter
	filter := make(map[string]interface{})
	_, err := DB.ListCollectionNames(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	return nil
}

// CloseDatabase gracefully closes the MongoDB connection
func CloseDatabase() error {
	if MongoClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := MongoClient.Disconnect(ctx)
	if err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("Successfully disconnected from MongoDB")
	return nil
}

// HealthCheck performs a health check on the database connection
func HealthCheck() error {
	if MongoClient == nil {
		return fmt.Errorf("database client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := MongoClient.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
