# Redis-Go - Implémentation Redis en Go

Une implémentation minimale de Redis en Go avec les fonctionnalités de base.

## 🚀 Démarrage rapide

### Prérequis
- Go 1.24 ou plus récent

### Installation et lancement
```bash
# Cloner le projet
git clone https://github.com/yourname/redis-go.git
cd redis-go

# Initialiser les modules Go
go mod tidy

# Lancer le serveur
go run main.go
```

Le serveur démarre par défaut sur `localhost:6379`.

### Variables d'environnement
```bash
export REDIS_HOST=localhost        # Adresse d'écoute (défaut: localhost)
export REDIS_PORT=6379            # Port d'écoute (défaut: 6379)
export REDIS_MAX_CONNECTIONS=1000 # Nombre max de connexions (défaut: 1000)
export REDIS_EXPIRATION_CHECK_INTERVAL=1 # Intervalle GC en secondes (défaut: 1)
```

## 🛠️ Utilisation

### Connexion avec redis-cli
```bash
# Si vous avez redis-cli installé
redis-cli -h localhost -p 6379

# Ou avec telnet
telnet localhost 6379
```

### Commandes supportées

#### Commandes String
- `SET key value [EX seconds]` - Stocke une valeur avec TTL optionnel
- `GET key` - Récupère une valeur
- `DEL key [key ...]` - Supprime une ou plusieurs clés
- `EXISTS key [key ...]` - Vérifie l'existence de clés

#### Commandes utilitaires
- `PING [message]` - Test de connexion
- `ECHO message` - Retourne le message
- `KEYS *` - Liste toutes les clés (pattern matching non implémenté)
- `DBSIZE` - Nombre de clés dans la base

### Exemples d'utilisation
```bash
# Stockage et récupération basique
SET mykey "Hello World"
GET mykey

# Avec expiration (10 secondes)
SET session:123 "user_data" EX 10
GET session:123

# Suppression multiple
DEL key1 key2 key3

# Vérification d'existence
EXISTS mykey
```

## 🏗️ Architecture

### Structure du projet
```
redis-go/
├── main.go                    # Point d'entrée
├── internal/
│   ├── config/               # Configuration
│   │   └── config.go
│   ├── storage/              # Stockage en mémoire
│   │   └── storage.go
│   ├── protocol/             # Protocole RESP
│   │   └── resp.go
│   ├── commands/             # Gestionnaire de commandes
│   │   └── commands.go
│   └── server/               # Serveur TCP
│       └── server.go
├── go.mod
└── README.md
```

### Composants principaux

#### 1. Storage (`internal/storage`)
- **Stockage clé-valeur en mémoire** avec `map[string]*Value`
- **Gestion de la concurrence** avec `sync.RWMutex`
- **Support des TTL** avec vérification d'expiration
- **Nettoyage lazy** : suppression à la lecture des clés expirées

#### 2. Protocol (`internal/protocol`)
- **Parser RESP** pour décoder les commandes clients
- **Encoder RESP** pour formater les réponses
- **Support complet** du protocole Redis (arrays, bulk strings, etc.)

#### 3. Commands (`internal/commands`)
- **Registry pattern** pour enregistrer les commandes
- **Validation des arguments** et gestion d'erreurs
- **Extensibilité** facile pour ajouter de nouvelles commandes

#### 4. Server (`internal/server`)
- **Serveur TCP multi-client** avec goroutines
- **Gestion des connexions** avec limite configurable
- **Garbage collector** automatique pour les clés expirées
- **Arrêt propre** avec gestion des signaux

### Choix techniques

#### Concurrence
- **Une goroutine par client** pour gérer les connexions simultanées
- **RWMutex sur le storage** : lectures simultanées, écritures exclusives
- **Channels pour la communication** entre composants

#### Expiration des clés
- **Lazy expiration** : vérification à la lecture (comme Redis)
- **Active expiration** : garbage collector périodique en arrière-plan
- **TTL stocké avec chaque valeur** pour éviter les index complexes

#### Protocole RESP
- **Parser streaming** avec `bufio.Reader` pour l'efficacité
- **Validation stricte** du format pour éviter les erreurs
- **Support des types principaux** (strings, integers, arrays, errors)

## ✅ Fonctionnalités implémentées

- [x] Serveur TCP avec connexions multiples
- [x] Protocole RESP (Redis Serialization Protocol)
- [x] Stockage clé-valeur en mémoire
- [x] Expiration automatique des clés (TTL)
- [x] Commandes String de base (SET, GET, DEL, EXISTS)
- [x] Commandes utilitaires (PING, ECHO, KEYS, DBSIZE)
- [x] Gestion propre des erreurs
- [x] Configuration par variables d'environnement
- [x] Garbage collector pour les clés expirées

## 🚧 Fonctionnalités manquantes (pour continuer le développement)

### Priorité haute
- [ ] **Types de données avancés** : Lists, Sets, Hashes, Sorted Sets
- [ ] **Persistence** : RDB snapshots et AOF (Append Only File)
- [ ] **Pattern matching** pour la commande KEYS
- [ ] **Commandes d'incrémentation** : INCR, DECR, INCRBY, DECRBY

### Priorité moyenne
- [ ] **Pub/Sub** : PUBLISH, SUBSCRIBE, UNSUBSCRIBE
- [ ] **Transactions** : MULTI, EXEC, DISCARD, WATCH
- [ ] **Commandes de configuration** : CONFIG GET/SET
- [ ] **Commandes d'information** : INFO, MONITOR

### Priorité basse
- [ ] **Clustering** et réplication
- [ ] **Scripting Lua**
- [ ] **Modules** et extensibilité
- [ ] **Compression** des données
- [ ] **Authentification** et sécurité

### Optimisations techniques
- [ ] **Index pour les TTL** (heap/priority queue) pour optimiser l'expiration
- [ ] **Pool de connections** pour réduire les allocations
- [ ] **Serialization binaire** plus efficace que les strings
- [ ] **Metrics et monitoring** intégrés

## 🔧 Comment reprendre le développement

### Pour ajouter un nouveau type de données (ex: Lists)

1. **Étendre `storage.DataType`**
```go
const (
    TypeList DataType = iota + 1 // Après les types existants
)
```

2. **Créer les structures de données**
```go
type RedisList struct {
    elements []string
    mutex    sync.RWMutex
}
```

3. **Ajouter les commandes dans `commands/`**
```go
r.commands["LPUSH"] = r.handleLpush
r.commands["RPUSH"] = r.handleRpush
r.commands["LPOP"] = r.handleLpop
// etc.
```

### Pour ajouter la persistence

1. **Créer un package `internal/persistence`**
2. **Implémenter RDB snapshots** (format binaire compact)
3. **Implémenter AOF** (log des commandes d'écriture)
4. **Ajouter la configuration** pour activer/désactiver
5. **Intégrer au serveur** avec des goroutines dédiées

### Pour ajouter Pub/Sub

1. **Créer `internal/pubsub`** avec gestion des abonnements
2. **Ajouter un canal de diffusion** dans le serveur
3. **Implémenter les commandes** PUBLISH, SUBSCRIBE, etc.
4. **Gérer les connexions persistantes** pour les subscribers

## 🧪 Tests

Pour tester le serveur :

```bash
# Test basique avec redis-cli
redis-cli -h localhost -p 6379 ping

# Test de performance simple
redis-cli -h localhost -p 6379 --latency-history -i 1

# Test avec script
redis-cli -h localhost -p 6379 eval "return 'Hello from Redis-Go'" 0
```

## 📝 Notes de développement

### Points d'attention pour la suite

1. **Gestion mémoire** : Attention au garbage collector Go avec de gros datasets
2. **Performance** : Profiler avec `go tool pprof` pour identifier les goulots
3. **Tests** : Ajouter des tests unitaires et d'intégration
4. **Documentation** : Documenter l'API interne pour faciliter les contributions

### Commandes utiles

```bash
# Profiling mémoire
go tool pprof http://localhost:6060/debug/pprof/heap

# Profiling CPU
go tool pprof http://localhost:6060/debug/pprof/profile

# Tests de charge
redis-benchmark -h localhost -p 6379 -q -n 100000
```

---

**État actuel** : MVP fonctionnel avec les bases de Redis
**Prochaine étape recommandée** : Implémenter les types Lists ou la persistence RDB