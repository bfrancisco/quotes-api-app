package runtime

import "testing"

func TestLoadConfigDefaultsToLocalMemory(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("STORAGE_MODE", "")
	t.Setenv("SEED_QUOTES", "")
	t.Setenv("FIRESTORE_PROJECT_ID", "")
	t.Setenv("FIRESTORE_DATABASE_ID", "")
	t.Setenv("FIRESTORE_COLLECTION", "")
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_SERVICE_VERSION", "")
	t.Setenv("DEPLOYMENT_ENVIRONMENT", "")

	config, err := LoadConfig("8080", "quotes-rest-api")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if config.Port != "8080" || config.StorageMode != StorageModeMemory || config.FirestoreCollection != "quotes" || config.SeedQuotes || config.TelemetryServiceName != "quotes-rest-api" {
		t.Fatalf("LoadConfig() = %#v, want local-memory defaults", config)
	}
}

func TestLoadConfigUsesTelemetryEnvironmentMetadata(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "custom-quotes-api")
	t.Setenv("OTEL_SERVICE_VERSION", "2026.08.17")
	t.Setenv("DEPLOYMENT_ENVIRONMENT", "staging")

	config, err := LoadConfig("8080", "quotes-rest-api")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if config.TelemetryServiceName != "custom-quotes-api" || config.TelemetryServiceVer != "2026.08.17" || config.DeploymentEnvironment != "staging" {
		t.Fatalf("LoadConfig() telemetry = %#v, want environment metadata", config)
	}
}

func TestLoadConfigRejectsInvalidFirestoreConfiguration(t *testing.T) {
	t.Setenv("STORAGE_MODE", StorageModeFirestore)
	t.Setenv("FIRESTORE_PROJECT_ID", "")

	if _, err := LoadConfig("8080", "quotes-rest-api"); err == nil {
		t.Fatal("LoadConfig() error = nil, want missing Firestore project error")
	}
}

func TestLoadConfigRejectsFirestoreSeeding(t *testing.T) {
	t.Setenv("STORAGE_MODE", StorageModeFirestore)
	t.Setenv("FIRESTORE_PROJECT_ID", "quotes-test")
	t.Setenv("SEED_QUOTES", "true")

	if _, err := LoadConfig("8080", "quotes-rest-api"); err == nil {
		t.Fatal("LoadConfig() error = nil, want Firestore seed configuration error")
	}
}
