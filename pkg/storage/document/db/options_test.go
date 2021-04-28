package db

import (
	"github.com/NerdShoreDev/YEP/server/pkg/srv"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestInitOptions(t *testing.T) {
	//arrange
	a := assert.New(t)
	env := srv.ServerValues{
		DBUser:            "Dark Helmet",
		DBPassword:        "12345",
		DBName:            "test-name",
		DBClusterEndpoint: "test.endpoint",
		DBConnectTimeout:  30,
		DBQueryTimeout:    20,
	}
	//act
	options := *NewOptions(&env)
	//assert
	a.Equal("Dark Helmet", options.User)
	a.Equal("12345", options.Password)
	a.Equal("test-name", options.DataBaseName)
	a.Equal("test.endpoint:27017",
		options.ClusterEndpoint)
	a.Equal(time.Duration(30), options.ConnectTimeout)
	a.Equal(time.Duration(20), options.QueryTimeout)
	a.Equal("secondaryPreferred", options.ReadPreference)
}

func TestBuildConnectionString(t *testing.T) {
	//arrange
	a := assert.New(t)
	env := srv.ServerValues{
		DBUser:            "BummBumm",
		DBPassword:        "12345",
		DBName:            "test-name",
		DBClusterEndpoint: "test.endpoint",
	}
	//act
	options := *NewOptions(&env)
	//assert
	a.Equal("mongodb://BummBumm:12345@test.endpoint/test-name?", options.ConnectionString)
}
