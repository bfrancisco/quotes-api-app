package runtime

import "testing"

func TestLoadConfigDefaultsToLocalMemory(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("STORAGE_MODE", "")
	t.Setenv("SEED_QUOTES", "")
	t.Setenv("FIRESTORE_PROJECT_ID", "")
	t.Setenv("FIRESTORE_DATABASE_ID", "")
	t.Setenv("FIRESTORE_COLLECTION", "")

	config, err := LoadConfig("8080")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if config.Port != "8080" || config.StorageMode != StorageModeMemory || config.FirestoreCollection != "quotes" || config.SeedQuotes {
		t.Fatalf("LoadConfig() = %#v, want local-memory defaults", config)
	}
}

func TestLoadConfigRejectsInvalidFirestoreConfiguration(t *testing.T) {
	t.Setenv("STORAGE_MODE", StorageModeFirestore)
	t.Setenv("FIRESTORE_PROJECT_ID", "")

	if _, err := LoadConfig("8080"); err == nil {
		t.Fatal("LoadConfig() error = nil, want missing Firestore project error")
	}
}

func TestLoadConfigRejectsFirestoreSeeding(t *testing.T) {
	t.Setenv("STORAGE_MODE", StorageModeFirestore)
	t.Setenv("FIRESTORE_PROJECT_ID", "quotes-test")
	t.Setenv("SEED_QUOTES", "true")

	if _, err := LoadConfig("8080"); err == nil {
		t.Fatal("LoadConfig() error = nil, want Firestore seed configuration error")
	}
}
