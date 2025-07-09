package storage

import (
	"fmt"
	"testing"
	"time"
)

// TestBasicOperations teste les opérations de base du storage
func TestBasicOperations(t *testing.T) {
	store := New()

	// Test Set/Get
	key := "test_key"
	value := "test_value"

	store.Set(key, value, TypeString, nil)

	retrieved := store.Get(key)
	if retrieved == nil {
		t.Error("Expected value to be retrieved, got nil")
	}

	if retrieved.Data.(string) != value {
		t.Errorf("Expected %s, got %s", value, retrieved.Data.(string))
	}

	if retrieved.Type != TypeString {
		t.Errorf("Expected type %d, got %d", TypeString, retrieved.Type)
	}
}

// TestExpiration teste la fonctionnalité d'expiration
func TestExpiration(t *testing.T) {
	store := New()

	key := "expiring_key"
	value := "expiring_value"
	ttl := 100 * time.Millisecond

	// Set avec TTL court
	store.Set(key, value, TypeString, &ttl)

	// Vérification immédiate
	retrieved := store.Get(key)
	if retrieved == nil {
		t.Error("Expected value to exist immediately after set")
	}

	// Attendre l'expiration
	time.Sleep(150 * time.Millisecond)

	// Vérification après expiration
	expired := store.Get(key)
	if expired != nil {
		t.Error("Expected value to be expired, but still exists")
	}
}

// TestDelete teste la suppression de clés
func TestDelete(t *testing.T) {
	store := New()

	key := "delete_test"
	value := "to_be_deleted"

	store.Set(key, value, TypeString, nil)

	// Vérification que la clé existe
	if !store.Exists(key) {
		t.Error("Key should exist before deletion")
	}

	// Suppression
	deleted := store.Delete(key)
	if !deleted {
		t.Error("Delete should return true for existing key")
	}

	// Vérification que la clé n'existe plus
	if store.Exists(key) {
		t.Error("Key should not exist after deletion")
	}

	// Tentative de suppression d'une clé inexistante
	deletedAgain := store.Delete(key)
	if deletedAgain {
		t.Error("Delete should return false for non-existing key")
	}
}

// TestConcurrency teste les accès concurrents (basique)
func TestConcurrency(t *testing.T) {
	store := New()

	// Test simple de concurrence : plusieurs goroutines qui écrivent/lisent
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			key := fmt.Sprintf("concurrent_key_%d", id)
			value := fmt.Sprintf("concurrent_value_%d", id)

			store.Set(key, value, TypeString, nil)
			retrieved := store.Get(key)

			if retrieved == nil || retrieved.Data.(string) != value {
				t.Errorf("Concurrent operation failed for key %s", key)
			}

			done <- true
		}(i)
	}

	// Attendre que toutes les goroutines terminent
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestCleanupExpired teste le nettoyage des clés expirées
func TestCleanupExpired(t *testing.T) {
	store := New()

	// Ajouter plusieurs clés avec des TTL courts
	ttl := 50 * time.Millisecond
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("cleanup_key_%d", i)
		store.Set(key, "value", TypeString, &ttl)
	}

	// Ajouter une clé sans TTL
	store.Set("permanent_key", "permanent_value", TypeString, nil)

	// Vérifier que toutes les clés existent
	if store.Size() != 6 {
		t.Errorf("Expected 6 keys, got %d", store.Size())
	}

	// Attendre l'expiration
	time.Sleep(100 * time.Millisecond)

	// Nettoyer les clés expirées
	cleaned := store.CleanupExpired()

	if cleaned != 5 {
		t.Errorf("Expected 5 cleaned keys, got %d", cleaned)
	}

	// Vérifier qu'il ne reste que la clé permanente
	if store.Size() != 1 {
		t.Errorf("Expected 1 key remaining, got %d", store.Size())
	}

	if !store.Exists("permanent_key") {
		t.Error("Permanent key should still exist")
	}
}
