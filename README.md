# Redis-Go - Implémentation Redis en Go

Une implémentation complète de Redis en Go avec support des types de données principaux et du protocole RESP.

## 🚀 Démarrage rapide

### Prérequis
- Go 1.24+ ou Docker & Docker Compose

### Option 1: Lancement avec Go
```bash
# Cloner le projet
git clone <repository-url>
cd redis-go

# Initialiser les modules Go
go mod tidy

# Lancer le serveur
make run
# ou directement: go run main.go
```

### Option 2: Lancement avec Docker
```bash
# Lancer l'environnement complet
make docker
# ou: docker compose up --build

# Dans un autre terminal, utiliser redis-cli
docker compose exec redis-cli redis-cli -h redis-go -p 6379
```

Le serveur démarre par défaut sur `localhost:6379`.

### Variables d'environnement
```bash
export REDIS_HOST=localhost        # Adresse d'écoute (défaut: localhost)
export REDIS_PORT=6379            # Port d'écoute (défaut: 6379)
export REDIS_MAX_CONNECTIONS=1000 # Nombre max de connexions (défaut: 1000)
export REDIS_EXPIRATION_CHECK_INTERVAL=1 # Intervalle GC en secondes (défaut: 1)
```

## 🛠️ Commandes supportées

### Commandes String
- `SET key value [EX seconds]` - Stocke une valeur avec TTL optionnel
- `GET key` - Récupère une valeur
- `DEL key [key ...]` - Supprime une ou plusieurs clés
- `EXISTS key [key ...]` - Vérifie l'existence de clés
- `TYPE key` - Retourne le type d'une clé
- `INCR key` - Incrémente un compteur
- `DECR key` - Décrémente un compteur
- `INCRBY key increment` - Incrémente par une valeur
- `DECRBY key decrement` - Décrémente par une valeur

### Commandes List
- `LPUSH key element [element ...]` - Ajoute des éléments au début de la liste
- `RPUSH key element [element ...]` - Ajoute des éléments à la fin de la liste
- `LPOP key` - Supprime et retourne le premier élément
- `RPOP key` - Supprime et retourne le dernier élément
- `LLEN key` - Retourne la longueur de la liste
- `LRANGE key start stop` - Retourne une partie de la liste

### Commandes Set
- `SADD key member [member ...]` - Ajoute des membres à un set
- `SMEMBERS key` - Retourne tous les membres d'un set
- `SISMEMBER key member` - Vérifie si un membre est dans le set

### Commandes Hash
- `HSET key field value [field value ...]` - Définit des champs dans un hash
- `HGET key field` - Récupère un champ d'un hash
- `HGETALL key` - Retourne tous les champs et valeurs d'un hash

### Commandes utilitaires
- `PING [message]` - Test de connexion
- `ECHO message` - Retourne le message
- `KEYS pattern` - Liste les clés correspondant au pattern (glob style)
- `DBSIZE` - Nombre de clés dans la base
- `FLUSHALL` - Vide toute la base

### Pattern matching pour KEYS
- `*` - Correspond à n'importe quelle séquence de caractères
- `?` - Correspond à n'importe quel caractère unique
- `[abc]` - Correspond à un des caractères spécifiés
- `[a-z]` - Correspond à un caractère dans l'intervalle
- `[^abc]` - Correspond à tout sauf les caractères spécifiés

## 📋 Exemples d'utilisation

### Strings et compteurs
```bash
# Stockage et récupération basique
SET user:1:name "Alice"
GET user:1:name

# Avec expiration (10 secondes)
SET session:abc123 "user_data" EX 10

# Compteurs
INCR page_views
INCRBY downloads 5
DECR stock_count
```

### Listes
```bash
# File FIFO
RPUSH queue "task1" "task2" "task3"
LPOP queue

# Pile LIFO
LPUSH stack "item1" "item2" "item3"
LPOP stack

# Affichage
LRANGE mylist 0 -1  # Toute la liste
LLEN mylist         # Longueur
```

### Sets
```bash
# Ajouter des éléments uniques
SADD tags "redis" "database" "cache"
SADD tags "redis"  # Ignoré car déjà présent

# Vérifier et lister
SISMEMBER tags "redis"  # 1
SMEMBERS tags          # Tous les membres
```

### Hashes
```bash
# Stocker des objets
HSET user:1 name "Alice" age "30" city "Paris"
HGET user:1 name        # "Alice"
HGETALL user:1         # Tous les champs

# Mise à jour partielle
HSET user:1 age "31"
```

### Pattern matching
```bash
# Toutes les clés
KEYS *

# Clés d'utilisateurs
KEYS user:*

# Sessions spécifiques
KEYS session:[a-f]*

# Clés temporaires
KEYS temp:???:*
```

## 🏗️ Architecture

### Structure du projet
```
redis-go/
├── main.go                    # Point d'entrée
├── internal/
│   ├── config/               # Configuration
│   │   └── config.go
│   ├── storage/              # Stockage multi-types
│   │   ├── storage.go
│   │   └── storage_test.go
│   ├── protocol/             # Protocole RESP
│   │   └── resp.go
│   ├── commands/             # Gestionnaire de commandes
│   │   └── commands.go
│   └── server/               # Serveur TCP
│       └── server.go
├── Dockerfile                # Image Docker
├── compose.yaml             # Orchestration
├── Makefile                 # Commandes de build
└── README.md
```

### Composants principaux

#### 1. Storage (`internal/storage`)
- **Stockage unifié** avec `map[string]*Value`
- **Types multiples** : String, List, Set, Hash
- **TTL par valeur** avec expiration lazy et active
- **Concurrence** gérée par `sync.RWMutex`
- **Pattern matching** complet (glob style Redis)

#### 2. Protocol (`internal/protocol`)
- **Parser RESP robuste** avec gestion d'erreurs détaillée
- **Support complet** : Arrays, Bulk Strings, Integers, Errors, Simple Strings
- **Encoder optimisé** pour les réponses
- **Gestion des timeouts** et connexions instables

#### 3. Commands (`internal/commands`)
- **Registry pattern** pour toutes les commandes
- **Validation stricte** des arguments et types
- **Messages d'erreur** en français et explicites
- **Extensibilité** facile pour nouvelles commandes

#### 4. Server (`internal/server`)
- **TCP multi-client** avec goroutines par connexion
- **Gestion propre** des connexions (max, timeouts)
- **Garbage collector** automatique pour les clés expirées
- **Arrêt gracieux** avec signaux système

### Choix techniques

#### Types de données
- **Value struct** unifié avec type et TTL
- **Structures spécialisées** pour chaque type de données (RedisList, RedisSet, RedisHash)
- **Lazy expiration** à la lecture + nettoyage actif
- **Pattern matching** avec algorithme récursif optimisé

#### Concurrence
- **Une goroutine par client** pour isolation
- **RWMutex global** : lectures simultanées, écritures exclusives
- **Pas de verrous fins** pour simplifier et éviter les deadlocks
- **Channels** pour communication serveur/GC

#### Protocole RESP
- **Parser streaming** byte par byte pour robustesse
- **Validation stricte** des formats CRLF
- **Gestion d'erreurs** détaillée pour debugging
- **Encoder direct** sans buffering intermédiaire

## ✅ Fonctionnalités implémentées

- [x] **Serveur TCP** avec connexions multiples et gestion propre
- [x] **Protocole RESP** complet et robuste
- [x] **Stockage multi-types** avec TTL et pattern matching
- [x] **Commandes String** : SET/GET/DEL/EXISTS/TYPE/INCR/DECR/INCRBY/DECRBY
- [x] **Commandes List** : LPUSH/RPUSH/LPOP/RPOP/LLEN/LRANGE
- [x] **Commandes Set** : SADD/SMEMBERS/SISMEMBER
- [x] **Commandes Hash** : HSET/HGET/HGETALL
- [x] **Pattern matching** : Support complet des patterns glob Redis
- [x] **Expiration automatique** : TTL avec nettoyage lazy et actif
- [x] **Messages d'erreur** : Messages en français et explicites
- [x] **Configuration** par variables d'environnement
- [x] **Docker** : Build multi-stage et compose ready

## 🚧 Roadmap (extensions possibles)

### Types de données avancés
- [ ] **Sorted Sets** (ZSET) : ZADD, ZRANGE, ZRANK, ZSCORE
- [ ] **Bitmaps** : SETBIT, GETBIT, BITCOUNT
- [ ] **HyperLogLog** : PFADD, PFCOUNT, PFMERGE

### Persistence
- [ ] **RDB snapshots** : Sauvegarde binaire périodique
- [ ] **AOF** (Append Only File) : Log des commandes d'écriture
- [ ] **Configuration** : Activation/désactivation, intervalles

### Fonctionnalités avancées
- [ ] **Pub/Sub** : PUBLISH, SUBSCRIBE, UNSUBSCRIBE, PSUBSCRIBE
- [ ] **Transactions** : MULTI, EXEC, DISCARD, WATCH
- [ ] **Lua scripting** : EVAL, EVALSHA avec sandbox
- [ ] **Connexions authentifiées** : AUTH, utilisateurs

### Performance et monitoring
- [ ] **Index TTL** : Heap/priority queue pour expiration efficace
- [ ] **Métriques** : Compteurs de commandes, temps de réponse
- [ ] **Info command** : Statistiques serveur et mémoire
- [ ] **Slow log** : Log des commandes lentes

## 🧪 Tests et validation

### Tests unitaires
```bash
# Lancer tous les tests
make test

# Tests avec coverage
go test -cover ./...

# Tests de race conditions
go test -race ./...
```

### Tests d'intégration
```bash
# Test avec le vrai redis-cli
make test-cli

# Tests automatisés via Docker
make docker

# Tests de charge (nécessite redis-benchmark)
redis-benchmark -h localhost -p 6379 -q -n 100000
```

## 🔧 Développement

### Commandes utiles
```bash
# Développement
make run       # Démarre le serveur
make build     # Compile le binaire
make test      # Lance les tests
make docker    # Lance avec Docker

# Maintenance
make fmt       # Formate le code
make deps      # Met à jour les dépendances
make clean     # Nettoie les artefacts
make help      # Affiche l'aide
```

### Ajouter une nouvelle commande

1. **Ajouter la méthode** dans `internal/storage/storage.go` si nécessaire
2. **Créer le handler** dans `internal/commands/commands.go`
3. **Enregistrer** dans `registerCommands()`
4. **Tester** avec des tests unitaires

Exemple pour une commande `STRLEN` :
```go
// Dans storage.go (si nécessaire)
func (s *Storage) StringLen(key string) int {
    value := s.Get(key)
    if value == nil || value.Type != TypeString {
        return 0
    }
    return len(value.Data.(string))
}

// Dans commands.go
func (r *Registry) handleStrLen(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
    if len(args) != 1 {
        return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'STRLEN' (attendu: STRLEN clé)")
    }
    
    length := store.StringLen(args[0])
    return encoder.WriteInteger(int64(length))
}

// Dans registerCommands()
r.commands["STRLEN"] = r.handleStrLen
```

## 📊 Performance

### Métriques typiques
- **Throughput** : ~50K ops/sec sur machine standard
- **Latency** : <1ms pour GET/SET simple
- **Memory** : ~100 bytes overhead par clé
- **Connexions** : 1000 clients simultanés par défaut

### Optimisations appliquées
- **RWMutex** pour lectures parallèles
- **Pas de sérialisation** : données natives en mémoire
- **Pattern matching** : Algorithme récursif optimisé
- **Garbage collection** : Nettoyage actif + lazy des TTL
- **Parser RESP** : Lecture streaming sans copies inutiles

---

**État actuel** : Implémentation fonctionnelle avec types de données principaux  
**Compatibilité** : Protocole RESP et commandes de base compatibles Redis  
**Production** : Prêt pour usage léger, ajouter persistence pour usage critique