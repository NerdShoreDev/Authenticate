package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"time"
)

// ClientWrapper acts as a database wrapper
type DatabaseWrapper struct {
	Database     *mongo.Database
	QueryTimeout time.Duration
}

// NewDatabaseWrapper is a method that constructs a new databaseWrapper
func NewDatabaseWrapper(dbOptions *Options) *DatabaseWrapper {
	ctx, cancel := context.WithTimeout(context.Background(), dbOptions.ConnectTimeout*time.Second)
	defer cancel()

	client, err := connect(dbOptions, ctx)
	if err != nil {
		log.Fatalf("Failed to connect to cluster: %v", err)
	}
	// defer client.Disconnect(ctx)

	// Force a connection to verify our connection string
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping cluster: %v", err)
	}

	log.Print("Connected to DocumentDB!")

	database := client.Database(dbOptions.DataBaseName)

	return &DatabaseWrapper{
		Database:     database,
		QueryTimeout: dbOptions.QueryTimeout,
	}
}

func connect(dbOptions *Options, ctx context.Context) (*mongo.Client, error) {
	if dbOptions.UseSSL != "true" {
		return mongo.Connect(ctx, options.Client().ApplyURI(dbOptions.ConnectionString))
	}
	tlsConfig, err := getCustomTLSConfig(dbOptions.CaFilePath)
	if err != nil {
		return nil, err
	}
	return mongo.Connect(ctx, options.Client().ApplyURI(dbOptions.ConnectionString).SetTLSConfig(tlsConfig))
}

func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()

	if ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs); !ok {
		return tlsConfig, errors.New("failed parsing pem file")
	}

	return tlsConfig, nil
}
