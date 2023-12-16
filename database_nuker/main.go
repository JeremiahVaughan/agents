package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strings"
	"time"
)

var (
	RedisConnectionPool     *redis.Client
	CockroachConnectionPool *pgxpool.Pool
)

func main() {
	ctx := context.Background()
	err := establishConnections(ctx)
	if err != nil {
		log.Fatalf("error, when establishConnections() for main(). Error: %v", err)
	}

	err = RedisConnectionPool.FlushDB(ctx).Err()
	if err != nil {
		log.Fatalf("error, when FlushDB() for main(). Error: %v", err)
	}

	dbName, err := getDatabaseName()
	if err != nil {
		log.Fatalf("error, when getDatabaseName() for main(). Error: %v", err)
	}
	_, err = CockroachConnectionPool.Exec(
		ctx,
		"DROP DATABASE IF EXISTS $1 CASCADE",
		dbName,
	)
	if err != nil {
		log.Fatalf("error, when attempting to drop Cockroach database. Error: %v", err)
	}

	_, err = CockroachConnectionPool.Exec(
		ctx,
		"CREATE DATABASE $1",
		dbName,
	)
	if err != nil {
		log.Fatalf("error, when attempting to recreate the Cockroach database. Error: %v", err)
	}
}

func getDatabaseName() (string, error) {
	databaseName := os.Getenv("TF_VAR_database_name")
	if databaseName == "" {
		return "", fmt.Errorf("error, must provide env var: TF_VAR_database_name")
	}
	return databaseName, nil
}

func establishConnections(ctx context.Context) error {

	redisConnectionString, redisPassword, err := getRedisCredentials()
	if err != nil {
		return fmt.Errorf("error, when getRedisCredentials() for establishConnections(). Error: %v", err)
	}
	RedisConnectionPool, err = connectToRedisDatabase(redisConnectionString, redisPassword)
	if err != nil {
		return fmt.Errorf("error, when connectToRedisDatabase() for establishConnections(). Error: %v", err)
	}
	_, err = RedisConnectionPool.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("error, when attempting to ping the redis database after establishing the initial pooling connection: %v", err)
	}

	databaseConnectionString, databaseRootCa, err := getCockroachCredentials()
	if err != nil {
		return fmt.Errorf("error, when getCockroachCredentials() for establishConnections(). Error: %v", err)
	}
	CockroachConnectionPool, err = connectToCockroachDatabase(ctx, databaseConnectionString, databaseRootCa)
	if err != nil {
		return fmt.Errorf("error, when attempting to establish a connection pool with the database: %v", err)
	}
	err = CockroachConnectionPool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("error, when attempting to ping databse after establishing the initial pooling connection with CockroachDB. Error: %v", err)
	}

	return nil
}

func getCockroachCredentials() (string, string, error) {
	var errorMsgs []string
	databaseConnectionString := os.Getenv("TF_VAR_database_connection_string")
	if databaseConnectionString == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_database_connection_string")
	}
	databaseRootCa := os.Getenv("TF_VAR_database_root_ca")
	if databaseRootCa == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_database_root_ca")
	}
	if len(errorMsgs) > 0 {
		return "", "", fmt.Errorf("error, missing environment variables: %v", strings.Join(errorMsgs, ", "))
	}
	return databaseConnectionString, databaseRootCa, nil
}

func connectToCockroachDatabase(ctx context.Context, connectionString, databaseRootCa string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("error, when parsing connection string: %v", err)
	}

	var tlsConfig *tls.Config
	tlsConfig, err = generateCockroachTlsConfig(databaseRootCa, config)
	if err != nil {
		return nil, fmt.Errorf("error, when generating tls config: %v", err)
	}
	// Add the TLS configuration to the connection config
	config.ConnConfig.TLSConfig = tlsConfig

	// Customize connection pool settings (if desired)
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Minute * 30
	config.MaxConnIdleTime = time.Minute * 5
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("error, unable to create connection pool: %v", err)
	}

	err = attemptToPingCockroachDatabaseUntilSuccessful(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("error, exausted attempts to ping the database: %v", err)
	}

	return pool, nil
}

func attemptToPingCockroachDatabaseUntilSuccessful(ctx context.Context, pool *pgxpool.Pool) error {
	timeOutInSeconds := 45
	retryInterval := 3
	var err error
	for i := 0; i < (timeOutInSeconds / retryInterval); i++ {
		err = pool.Ping(ctx)
		if err != nil {
			log.Printf("Database ping failed, will attempt again in: %d seconds...", retryInterval)
			time.Sleep(time.Duration(retryInterval) * time.Second)
		} else {
			break
		}
	}
	return err
}

func generateCockroachTlsConfig(databaseRootCa string, config *pgxpool.Config) (*tls.Config, error) {
	rootCAs, err := loadCockroachRootCA(databaseRootCa)
	if err != nil {
		return nil, fmt.Errorf("error, failed to load root CA for generateTlsConfig(): %v", err)
	}
	tlsConfig := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: config.ConnConfig.Host,
	}
	return tlsConfig, nil
}

func loadCockroachRootCA(databaseRootCa string) (*x509.CertPool, error) {
	var err error
	var decodedCert []byte

	// The cert is encoded when deployed because I need to pass it around with terraform
	decodedCert, err = base64.StdEncoding.DecodeString(databaseRootCa)
	if err != nil {
		return nil, fmt.Errorf("error, when base64 decoding database CA cert: %v", err)
	}

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(decodedCert); !ok {
		return nil, fmt.Errorf("error, failed to append CA certificate to the certificate pool")
	}
	return rootCAs, nil
}

func getRedisCredentials() (string, string, error) {
	var errorMsgs []string
	redisConnectionString := os.Getenv("TF_VAR_redis_connection_string")
	if redisConnectionString == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_redis_connection_string")
	}

	redisPassword := os.Getenv("TF_VAR_redis_password")
	if redisPassword == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_redis_password")
	}

	if len(errorMsgs) > 0 {
		return "", "", fmt.Errorf("error, missing environment variables: %v", strings.Join(errorMsgs, ", "))
	}
	return redisConnectionString, redisPassword, nil
}

func connectToRedisDatabase(connectionString string, password string) (*redis.Client, error) {
	options := redis.Options{
		Addr:     connectionString,
		DB:       0, // use default DB
		Password: password,
	}
	// Load client cert
	clientCert, clientKey, err := getRedisClientCertAndKey()
	if err != nil {
		return nil, fmt.Errorf("error, when getClientCertAndKey() for connectToRedisDatabase(). Error: %v", err)
	}
	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting set LoadX509KeyPair() for connectToRedisDatabase(). Error: %v", err) // Ye don't want to sail without a map!
	}

	caCert, err := getRedisCaCert()
	if err != nil {
		return nil, fmt.Errorf("error, when getCaCert() for connectToRedisDatabase(). Error: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	options.TLSConfig = tlsConfig
	return redis.NewClient(&options), nil
}

func getRedisClientCertAndKey() ([]byte, []byte, error) {
	clientCert, err := base64.StdEncoding.DecodeString(os.Getenv("TF_VAR_redis_user_crt"))
	if err != nil {
		return nil, nil, fmt.Errorf("error, when decoding clientCert. Error: %v", err)
	}

	clientKey, err := base64.StdEncoding.DecodeString(os.Getenv("TF_VAR_redis_user_private_key"))
	if err != nil {
		return nil, nil, fmt.Errorf("error, when decoding clientKey. Error: %v", err)
	}

	return clientCert, clientKey, nil
}

func getRedisCaCert() ([]byte, error) {
	result, err := base64.StdEncoding.DecodeString(os.Getenv("TF_VAR_redis_ca"))
	if err != nil {
		return nil, fmt.Errorf("error, when decoding CA cert. Error: %v", err)
	}
	return result, nil
}
