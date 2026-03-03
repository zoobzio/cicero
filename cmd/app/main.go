// Package main is the entry point for the public API binary.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/zoobzio/aperture"
	"github.com/zoobzio/astql/postgres"
	"github.com/zoobzio/capitan"
	"github.com/zoobzio/pipz"
	"github.com/zoobzio/sum"

	"github.com/zoobzio/cicero/api/contracts"
	"github.com/zoobzio/cicero/api/handlers"
	"github.com/zoobzio/cicero/api/wire"
	"github.com/zoobzio/cicero/config"
	"github.com/zoobzio/cicero/events"
	extranslator "github.com/zoobzio/cicero/external/translator"
	intotel "github.com/zoobzio/cicero/internal/otel"
	"github.com/zoobzio/cicero/internal/classify"
	"github.com/zoobzio/cicero/internal/translate"
	"github.com/zoobzio/cicero/models"
	"github.com/zoobzio/cicero/stores"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("starting...")
	ctx := context.Background()

	// Initialize sum service and registry.
	svc := sum.New()
	k := sum.Start()

	// =========================================================================
	// 1. Load Configuration
	// =========================================================================

	if err := sum.Config[config.App](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
	if err := sum.Config[config.Database](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}
	if err := sum.Config[config.Translator](ctx, k, nil); err != nil {
		return fmt.Errorf("failed to load translator config: %w", err)
	}

	// =========================================================================
	// 2. Connect to Infrastructure
	// =========================================================================

	dbCfg := sum.MustUse[config.Database](ctx)
	db, err := sqlx.Connect("postgres", dbCfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()
	log.Println("database connected")
	capitan.Emit(ctx, events.StartupDatabaseConnected)

	// =========================================================================
	// 3. Create Stores
	// =========================================================================

	renderer := postgres.New()
	allStores, err := stores.New(db, renderer)
	if err != nil {
		return fmt.Errorf("failed to create stores: %w", err)
	}

	// =========================================================================
	// 4. Create Clients and Services
	// =========================================================================

	translatorCfg := sum.MustUse[config.Translator](ctx)
	translatorClient := extranslator.NewClient(translatorCfg.Addr)
	defer func() { _ = translatorClient.Close() }()

	classifier := &classify.Simple{}

	// =========================================================================
	// 5. Register Contracts
	// =========================================================================

	sum.Register[contracts.Sources](k, allStores.Sources)
	sum.Register[contracts.Translations](k, allStores.Translations)
	sum.Register[contracts.Translator](k, translatorClient)
	sum.Register[classify.Classifier](k, classifier)

	// Register the translation pipeline so handlers can resolve it from context.
	pipeline := translate.NewPipeline()
	sum.Register[pipz.Chainable[*translate.Job]](k, pipeline)

	// =========================================================================
	// 6. Register Boundaries
	// =========================================================================

	if err := models.RegisterBoundaries(k); err != nil {
		return fmt.Errorf("failed to register model boundaries: %w", err)
	}
	if err := wire.RegisterBoundaries(k); err != nil {
		return fmt.Errorf("failed to register wire boundaries: %w", err)
	}

	// =========================================================================
	// 7. Freeze Registry
	// =========================================================================

	sum.Freeze(k)
	capitan.Emit(ctx, events.StartupServicesReady)

	// =========================================================================
	// 8. Initialize Observability (OTEL + Aperture)
	// =========================================================================

	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "localhost:4318"
	}
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "cicero"
	}

	otelProviders, err := intotel.New(ctx, intotel.Config{
		Endpoint:    otelEndpoint,
		ServiceName: serviceName,
	})
	if err != nil {
		return fmt.Errorf("failed to create otel providers: %w", err)
	}
	defer func() { _ = otelProviders.Shutdown(ctx) }()
	log.Println("observability initialized")
	capitan.Emit(ctx, events.StartupOTELReady)

	ap, err := aperture.New(
		capitan.Default(),
		otelProviders.Log,
		otelProviders.Metric,
		otelProviders.Trace,
	)
	if err != nil {
		return fmt.Errorf("failed to create aperture: %w", err)
	}
	defer ap.Close()
	capitan.Emit(ctx, events.StartupApertureReady)

	// =========================================================================
	// 9. Register Handlers and Run
	// =========================================================================

	svc.Handle(handlers.All()...)

	appCfg := sum.MustUse[config.App](ctx)
	capitan.Emit(ctx, events.StartupServerListening, events.StartupPortKey.Field(appCfg.Port))
	log.Printf("starting server on port %d...", appCfg.Port)
	return svc.Run("", appCfg.Port)
}
