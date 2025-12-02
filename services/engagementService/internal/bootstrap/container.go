package bootstrap

import (
	"context"
	"database/sql" // Import for PostgreSQL
	"engagementService/internal/config"
	ms "engagementService/internal/messaging"
	"engagementService/internal/repository"
	"engagementService/internal/service"
	"fmt"
	"github.com/Sayan80bayev/go-project/pkg/caching"
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/Sayan80bayev/go-project/pkg/messaging"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // File source for migrations
	_ "github.com/lib/pq"                                // PostgreSQL driver
	"time"
)

// Container holds all dependencies
type Container struct {
	DB                  *sql.DB // Changed from *mongo.Database to *sql.DB
	Redis               caching.CacheService
	Producer            messaging.Producer
	Consumer            messaging.Consumer
	SubscriptionService *service.SubscriptionService
	LikeService         *service.LikeService
	Config              *config.Config
	JWKSUrl             string
}

// Init initializes all dependencies and returns a container
func Init() (*Container, error) {
	logger := logging.GetLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize PostgreSQL database and run migrations
	db, err := initPostgresDatabase(cfg)
	if err != nil {
		return nil, err
	}

	// Run database migrations
	//if err := runMigrations(db, cfg); err != nil {
	//	return nil, fmt.Errorf("failed to run migrations: %w", err)
	//}

	cacheService, err := initRedis(cfg)
	if err != nil {
		return nil, err
	}

	producer, err := initRabbitMQProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Use the new PostgresSubscriptionRepo
	subRepo := repository.NewPostgresSubscriptionRepo(db) // Changed to NewPostgresSubscriptionRepo
	subService := service.NewSubscriptionService(subRepo, producer)

	likeRepo := repository.NewPostgresLikeRepo(db)
	likeService := service.NewLikeService(likeRepo)

	jwksURL := buildJWKSURL(cfg)

	logger.Info("Dependencies initialized successfully")

	return &Container{
		DB:                  db,
		Redis:               cacheService,
		Producer:            producer,
		Config:              cfg,
		JWKSUrl:             jwksURL,
		SubscriptionService: subService,
		LikeService:         likeService,
	}, nil
}

// --- Helpers ---

// initPostgresDatabase initializes a PostgreSQL database connection
func initPostgresDatabase(cfg *config.Config) (*sql.DB, error) {
	logger := logging.GetLogger()
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Ping the database to verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	logger.Info("PostgreSQL connected")
	return db, nil
}

// runMigrations applies database migrations from the migrations directory
func runMigrations(db *sql.DB, cfg *config.Config) error {
	logger := logging.GetLogger()

	// Initialize the PostgreSQL driver for migrations
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize migration driver: %w", err)
	}

	// Initialize the migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // Path to migrations directory
		"postgres",          // Database name (matches driver)
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	// Apply migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("Database migrations applied successfully")
	return nil
}

func initRedis(cfg *config.Config) (*caching.RedisService, error) {
	logger := logging.GetLogger()
	redisCache, err := caching.NewRedisService(caching.RedisConfig{
		DB:       0,
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
	})

	if err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	logger.Info("Redis connected")
	return redisCache, nil
}

func initRabbitMQProducer(cfg *config.Config) (messaging.Producer, error) {
	logger := logging.GetLogger()
	ampq := buildAmqpURL(cfg)
	prod, err := ms.NewRabbitProducer(ampq, cfg.RabbitMQExchange, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create AMQP producer: %w", err)
	}
	logger.Info("RabbitMQ producer created")

	return prod, nil
}

func buildJWKSURL(cfg *config.Config) string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", cfg.KeycloakURL, cfg.KeycloakRealm)
}

func buildAmqpURL(cfg *config.Config) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQUser,
		cfg.RabbitMQPassword,
		cfg.RabbitMQHost,
		cfg.RabbitMQPort,
	)
}
