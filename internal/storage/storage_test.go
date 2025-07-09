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

// TestPatternMatching teste le pattern matching pour KEYS
func TestPatternMatching(t *testing.T) {
	store := New()

	// Créer plusieurs clés de test
	testKeys := []string{
		"user:123",
		"user:456",
		"session:abc",
		"session:def",
		"cache:temp",
		"data",
	}

	for _, key := range testKeys {
		store.Set(key, "value", TypeString, nil)
	}

	// Tests de patterns
	testCases := []struct {
		pattern  string
		expected []string
	}{
		{"*", testKeys}, // Toutes les clés
		{"user:*", []string{"user:123", "user:456"}},
		{"session:*", []string{"session:abc", "session:def"}},
		{"*:*", []string{"user:123", "user:456", "session:abc", "session:def", "cache:temp"}},
		{"data", []string{"data"}},
		{"nonexistent", []string{}},
		{"user:123", []string{"user:123"}},
		{"*temp", []string{"cache:temp"}},
		{"????", []string{"data"}}, // 4 caractères
	}

	for _, tc := range testCases {
		result := store.Keys(tc.pattern)

		// Vérifier que tous les éléments attendus sont présents
		if len(result) != len(tc.expected) {
			t.Errorf("Pattern '%s': expected %d keys, got %d. Expected: %v, Got: %v",
				tc.pattern, len(tc.expected), len(result), tc.expected, result)
			continue
		}

		// Créer une map pour vérifier la présence
		expectedMap := make(map[string]bool)
		for _, key := range tc.expected {
			expectedMap[key] = true
		}

		for _, key := range result {
			if !expectedMap[key] {
				t.Errorf("Pattern '%s': unexpected key '%s' in result", tc.pattern, key)
			}
		}
	}
}

// TestPatternMatchingWithExpiry teste KEYS avec des clés expirées
func TestPatternMatchingWithExpiry(t *testing.T) {
	store := New()

	// Créer des clés avec et sans TTL
	ttl := 50 * time.Millisecond
	store.Set("temp:1", "value", TypeString, &ttl)
	store.Set("temp:2", "value", TypeString, &ttl)
	store.Set("permanent:1", "value", TypeString, nil)

	// Vérifier que toutes les clés sont trouvées initialement
	keys := store.Keys("*")
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys initially, got %d", len(keys))
	}

	// Attendre l'expiration
	time.Sleep(100 * time.Millisecond)

	// Vérifier que seule la clé permanente reste
	keys = store.Keys("*")
	if len(keys) != 1 || keys[0] != "permanent:1" {
		t.Errorf("Expected only 'permanent:1', got %v", keys)
	}
}

// TestGlobPatternMatching teste directement la fonction de pattern matching
func TestGlobPatternMatching(t *testing.T) {
	testCases := []struct {
		pattern  string
		str      string
		expected bool
	}{
		// Wildcards basiques
		{"*", "anything", true},
		{"*", "", true},
		{"hello*", "hello", true},
		{"hello*", "helloworld", true},
		{"hello*", "hi", false},
		{"*world", "helloworld", true},
		{"*world", "world", true},
		{"*world", "hello", false},

		// Question mark
		{"h?llo", "hello", true},
		{"h?llo", "hallo", true},
		{"h?llo", "hllo", false},
		{"h?llo", "helllo", false},

		// Caractères exacts
		{"hello", "hello", true},
		{"hello", "Hello", false},
		{"hello", "hell", false},

		// Classes de caractères
		{"[abc]", "a", true},
		{"[abc]", "b", true},
		{"[abc]", "d", false},
		{"test[123]", "test1", true},
		{"test[123]", "test4", false},

		// Classes négatives
		{"[^abc]", "d", true},
		{"[^abc]", "a", false},

		// Ranges
		{"[a-z]", "m", true},
		{"[a-z]", "A", false},
		{"[0-9]", "5", true},
		{"[0-9]", "a", false},

		// Combinaisons complexes
		{"user:[0-9]*", "user:123abc", true},
		{"user:[0-9]*", "user:abc123", false},
		{"*:[a-z][a-z][a-z]", "session:abc", true},
		{"*:[a-z][a-z][a-z]", "session:ab", false},
	}

	for _, tc := range testCases {
		result := matchGlobPattern(tc.pattern, tc.str)
		if result != tc.expected {
			t.Errorf("matchGlobPattern('%s', '%s') = %v, expected %v",
				tc.pattern, tc.str, result, tc.expected)
		}
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
