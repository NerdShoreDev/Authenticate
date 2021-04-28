package main

import (
	"github.com/NerdShoreDev/YEP/server/pkg/auth"
	moduleFactory "github.com/NerdShoreDev/YEP/server/pkg/module/factory"
	moduleHandler "github.com/NerdShoreDev/YEP/server/pkg/module/handler"
	moduleRepository "github.com/NerdShoreDev/YEP/server/pkg/module/repository"
	registryFactory "github.com/NerdShoreDev/YEP/server/pkg/registry/factory"
	registryHandler "github.com/NerdShoreDev/YEP/server/pkg/registry/handler"
	registryRepository "github.com/NerdShoreDev/YEP/server/pkg/registry/repository"
	"github.com/NerdShoreDev/YEP/server/pkg/service"
	"time"

	"github.com/NerdShoreDev/YEP/server/pkg/http/rest"
	"github.com/NerdShoreDev/YEP/server/pkg/srv"
	"github.com/NerdShoreDev/YEP/server/pkg/storage/document/db"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Load .env file if present
	envFile := ".env"
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Failed to load file '%s': %v", envFile, err)
	}

	// Initialize dependencies
	serverValues := srv.NewServerValues()

	// Set log level
	srv.SetLogLevel(serverValues.LogLevel, serverValues.Version)
	log.Debugln("serverValues ", serverValues)

	// Initialize Sentry
	if err := sentry.Init(sentry.ClientOptions{
		AttachStacktrace: true,
		Dsn:              serverValues.SentryClientKey,
		Environment:      serverValues.Environment,
	}); err != nil {
		log.Errorf("Sentry initialization failed: %v\n", err)
	}
	defer sentry.Flush(5 * time.Second)

	// Hook Sentry into logrus
	sentryHook := srv.NewSentryHook([]log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
	})
	log.AddHook(sentryHook)

	// Initialise DB connection and Database
	dbWrapper := db.NewDatabaseWrapper(db.NewOptions(serverValues))

	// Initialise Registry Storage, Factory and Handler
	registryRepository := registryRepository.NewRegistryStorage(dbWrapper.Database, *serverValues)
	registryFactory := registryFactory.NewRegistryFactory()

	// Initialise Modules Storage, Factory and Handler
	modulesRepository := moduleRepository.NewModulesStorage(dbWrapper.Database, *serverValues)
	modulesFactory := moduleFactory.NewModuleFactory()

	// Initialize handlers
	registryHandler := registryHandler.NewRegistryHandler(registryFactory, registryRepository, modulesRepository, serverValues)
	if err := registryHandler.InitRegistryData(); err != nil {
		log.Errorf("Unable to save registry data")
	}

	// Initialize rest client as validation server caller
	restClient := &rest.Client{
		AuthBaseUrl:       serverValues.AuthBaseUrl,
		ValidationBaseUrl: serverValues.ValidationBaseUrl,
	}

	modulesHandler := moduleHandler.NewModulesHandler(modulesFactory, modulesRepository, serverValues)

	// Initialize ServiceHandler & JWTHandler
	jwtHandler := auth.NewJwtHandler(restClient, serverValues.AuthTokenValidationIssuer, serverValues.AuthTokenValidationAudience)
	serviceHandler := service.NewService(jwtHandler, modulesHandler, registryHandler, restClient, serverValues)

	webServer := rest.NewWebServer(serverValues.AllowedOrigins)
	webServer.StartWebServer(serviceHandler)
}
