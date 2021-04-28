package db

import (
	"fmt"
	"time"

	"github.com/NerdShoreDev/YEP/server/pkg/srv"
)

// Options provide database options
type Options struct {
	User             string
	Password         string
	DataBaseName     string
	ClusterEndpoint  string
	CaFilePath       string
	ConnectTimeout   time.Duration
	QueryTimeout     time.Duration
	UseSSL           string
	ConnectionString string
	ReadPreference   string
}

// NewOptions instantiates a new Options object
func NewOptions(serverValues *srv.ServerValues) *Options {
	return &Options{
		User:             serverValues.DBUser,
		Password:         serverValues.DBPassword,
		DataBaseName:     serverValues.DBName,
		ClusterEndpoint:  serverValues.DBClusterEndpoint + ":27017",
		CaFilePath:       serverValues.DBCaFilePath,
		ConnectionString: buildConnectionString(serverValues),
		ConnectTimeout:   time.Duration(serverValues.DBConnectTimeout),
		UseSSL:           serverValues.DBSsl,
		QueryTimeout:     time.Duration(serverValues.DBQueryTimeout),
		ReadPreference:   "secondaryPreferred",
	}
}

func buildConnectionString(serverValues *srv.ServerValues) string {
	var dbPrefix = "mongodb"
	var authValues = ""
	var queryString = ""
	var hostString = fmt.Sprintf("%s/%s", serverValues.DBClusterEndpoint, serverValues.DBName)

	if len(serverValues.DBUser) > 0 || len(serverValues.DBPassword) > 0 {
		authValues = fmt.Sprintf("%s:%s@", serverValues.DBUser, serverValues.DBPassword)
	}

	if serverValues.DBSsl == "true" && len(serverValues.DBCaFilePath) > 0 {
		queryString += fmt.Sprintf("ssl=true&")
	}

	if len(serverValues.DBReplicaSet) > 0 {
		queryString += fmt.Sprintf("replicaSet=%s&readpreference=%s", serverValues.DBReplicaSet, "primaryPreferred")
	}

	return fmt.Sprintf("%s://%s%s?%s", dbPrefix, authValues, hostString, queryString)
}
