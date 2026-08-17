package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	StorageModeMemory    = "memory"
	StorageModeFirestore = "firestore"
)

// Config contains the runtime settings supplied by local environment variables or Cloud Run.
type Config struct {
	Port                  string
	StorageMode           string
	FirestoreProjectID    string
	FirestoreDatabaseID   string
	FirestoreCollection   string
	SeedQuotes            bool
	TelemetryServiceName  string
	TelemetryServiceVer   string
	DeploymentEnvironment string
}

// LoadConfig reads and validates runtime settings. The supplied defaultPort is used only when
// Cloud Run's PORT environment variable is not present. defaultTelemetryServiceName identifies
// the executable unless OTEL_SERVICE_NAME overrides it.
func LoadConfig(defaultPort, defaultTelemetryServiceName string) (Config, error) {
	config := Config{
		Port:                  valueOrDefault("PORT", defaultPort),
		StorageMode:           valueOrDefault("STORAGE_MODE", StorageModeMemory),
		FirestoreProjectID:    strings.TrimSpace(os.Getenv("FIRESTORE_PROJECT_ID")),
		FirestoreDatabaseID:   strings.TrimSpace(os.Getenv("FIRESTORE_DATABASE_ID")),
		FirestoreCollection:   valueOrDefault("FIRESTORE_COLLECTION", "quotes"),
		SeedQuotes:            parseBoolOrDefault("SEED_QUOTES", false),
		TelemetryServiceName:  valueOrDefault("OTEL_SERVICE_NAME", defaultTelemetryServiceName),
		TelemetryServiceVer:   strings.TrimSpace(os.Getenv("OTEL_SERVICE_VERSION")),
		DeploymentEnvironment: strings.TrimSpace(os.Getenv("DEPLOYMENT_ENVIRONMENT")),
	}

	if config.Port == "" {
		return Config{}, fmt.Errorf("PORT must not be empty")
	}
	if config.TelemetryServiceName == "" {
		return Config{}, fmt.Errorf("OTEL_SERVICE_NAME must not be empty")
	}
	if config.StorageMode != StorageModeMemory && config.StorageMode != StorageModeFirestore {
		return Config{}, fmt.Errorf("STORAGE_MODE must be %q or %q", StorageModeMemory, StorageModeFirestore)
	}
	if config.SeedQuotes && config.StorageMode != StorageModeMemory {
		return Config{}, fmt.Errorf("SEED_QUOTES=true is supported only with STORAGE_MODE=%q", StorageModeMemory)
	}
	if config.StorageMode == StorageModeFirestore && config.FirestoreProjectID == "" {
		return Config{}, fmt.Errorf("FIRESTORE_PROJECT_ID is required when STORAGE_MODE=%q", StorageModeFirestore)
	}

	return config, nil
}

func valueOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func parseBoolOrDefault(name string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
