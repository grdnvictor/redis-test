# Redis-Go
[![Go 1.24+](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker Ready](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://docker.com/)
[![badge de fatigue](https://img.shields.io/badge/On%20préfère%20le%20JS/TS%20pitié-💀-F7DF1E?style=flat&logo=typescript&logoColor=white&labelColor=3178C6)](https://www.typescriptlang.org/)
> Implémentation Redis partielle en Go avec protocole RESP (REdis Serialization Protocol) et support des types de données courants.

## 👨‍💻 Équipe de développement
|   **Développeur**    | **Classe** |
|:--------------------:|:----------:|
| **ALLARD Alexandre** | **`5IW2`** |
|  **GRANDIN Victor**  | **`5IW2`** |
|  **NKUMBA Estelle**  | **`5IW2`** |

## Fonctionnalités

### Types de données
- **Strings** avec TTL (INCR/DECR)
- **Lists** bidirectionnelles avec PUSH/POP
- **Sets** pour collections uniques
- **Hashes** pour objets structurés

### Protocole / Implémentation
- **RESP complet** compatible Redis
- **Pattern matching** avancé pour KEYS
- **Garbage collection** automatique des TTL

---

## Démarrage rapide

### Docker (Compose)

#### Démarrage
```bash
git clone <repository-url> && cd redis-go
make run
make cli  # Dans un autre terminal
```

#### Rédémarrage rapide (avec build)
```bash
make restart
```

### Go natif
```bash
go mod tidy && go run main.go
redis-cli -p 6379  # Test
```

### Premier test
```bash
SET welcome "Coucou Redis en GO !!" EX 3600
GET welcome
ALAIDE  # Voir toutes les commandes
```

---

## Architecture

### Structure du projet
```
redis-go/
├── main.go                    # Point d'entrée
├── internal/
│   ├── config/               # Configuration serveur
│   ├── protocol/             # Parser/Encoder RESP
│   ├── commands/             # Handlers de commandes
│   ├── storage/              # Moteur de stockage
│   └── server/               # Serveur TCP + lifecycle
├── Dockerfile                # Image Docker
├── compose.yml
└── Makefile                 # Commandes dev
```

### Flux de données

```mermaid
graph TB
    %% Clients
    CLI[redis-cli]
    APP[Applications]
    
    %% Serveur principal
    subgraph "🖥️ TCP Server"
        LISTENER[TCP Listener<br/>Port 6379]
        HANDLER[Client Handler<br/>Goroutine par client]
    end
    
    %% Protocole
    subgraph "🌐 RESP Protocol"
        PARSER[Command Parser<br/>Arrays → Args]
        ENCODER[Response Encoder<br/>Results → RESP]
    end
    
    %% Commandes
    subgraph "⚡ Command System"
        REGISTRY[Command Registry<br/>Dispatch & Validation]
        STRING_CMD[String Commands<br/>SET/GET/DEL]
        LIST_CMD[List Commands<br/>LPUSH/RPOP]
        SET_CMD[Set Commands<br/>SADD/SMEMBERS]
        HASH_CMD[Hash Commands<br/>HSET/HGET]
    end
    
    %% Stockage
    subgraph "💾 Storage Engine"
        CORE[Core Storage<br/>RWMutex + TTL]
        DATATYPES[Value Types<br/>String/List/Set/Hash]
        PATTERN[Pattern Matching<br/>Glob support]
    end
    
    %% Maintenance
    GC[Garbage Collector<br/>TTL cleanup]
    
    %% Flux principal
    CLI --> LISTENER
    APP --> LISTENER
    LISTENER --> HANDLER
    HANDLER --> PARSER
    PARSER --> REGISTRY
    
    REGISTRY --> STRING_CMD
    REGISTRY --> LIST_CMD
    REGISTRY --> SET_CMD
    REGISTRY --> HASH_CMD
    
    STRING_CMD --> CORE
    LIST_CMD --> CORE
    SET_CMD --> CORE
    HASH_CMD --> CORE
    
    CORE --> DATATYPES
    STRING_CMD --> PATTERN
    
    CORE --> ENCODER
    ENCODER --> HANDLER
    HANDLER --> CLI
    HANDLER --> APP
    
    GC --> CORE
    
    %% Styles
    classDef client fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef server fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef protocol fill:#e8f5e8,stroke:#388e3c,stroke-width:2px
    classDef command fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef storage fill:#fce4ec,stroke:#c2185b,stroke-width:2px
    classDef maintenance fill:#f1f8e9,stroke:#689f38,stroke-width:2px
    
    class CLI,APP client
    class LISTENER,HANDLER server
    class PARSER,ENCODER protocol
    class REGISTRY,STRING_CMD,LIST_CMD,SET_CMD,HASH_CMD command
    class CORE,DATATYPES,PATTERN storage
    class GC maintenance
```

---

## API des commandes

### Strings & Compteurs
| Commande | Syntaxe | Description |
|----------|---------|-------------|
| `SET` | `SET key value [EX seconds]` | Stocke avec TTL optionnel |
| `GET` | `GET key` | Récupère une valeur |
| `DEL` | `DEL key [key ...]` | Supprime des clés |
| `INCR` | `INCR key` | Incrémente de 1 |
| `INCRBY` | `INCRBY key increment` | Incrémente par N |

### Listes
| Commande | Syntaxe | Description |
|----------|---------|-------------|
| `LPUSH` | `LPUSH key element [element ...]` | Ajoute au début |
| `RPUSH` | `RPUSH key element [element ...]` | Ajoute à la fin |
| `LPOP` | `LPOP key` | Retire du début |
| `LLEN` | `LLEN key` | Longueur de liste |
| `LRANGE` | `LRANGE key start stop` | Sous-ensemble |

### Sets & Hashes
| Commande | Syntaxe | Description |
|----------|---------|-------------|
| `SADD` | `SADD key member [member ...]` | Ajoute des membres |
| `SMEMBERS` | `SMEMBERS key` | Liste tous les membres |
| `HSET` | `HSET key field value [field value ...]` | Définit des champs |
| `HGET` | `HGET key field` | Récupère un champ |

### Utilitaires
| Commande | Syntaxe | Description |
|----------|---------|-------------|
| `KEYS` | `KEYS pattern` | Recherche par motif (* ? [abc]) |
| `PING` | `PING [message]` | Test de connexion |
| `DBSIZE` | `DBSIZE` | Nombre de clés |
| `ALAIDE` | `ALAIDE [commande]` | Aide interactive |

---

## Configuration

### Variables d'environnement
```bash
REDIS_HOST=0.0.0.0              # Adresse d'écoute
REDIS_PORT=6379                 # Port du serveur
REDIS_MAX_CONNECTIONS=1000      # Connexions simultanées
REDIS_EXPIRATION_CHECK_INTERVAL=1  # GC interval (secondes)
```

### Docker Compose
```yaml
services:
  redis-go:
    build: .
    ports:
      - "6379:6379"
    environment:
      - REDIS_MAX_CONNECTIONS=2000
    restart: unless-stopped
```

---

## Développement

### Commandes Make
```bash
# Serveur
make help # Affiche les commandes disponibles !
```

## Exemples d'usage

### Cache avec expiration
```bash
SET cache:user:123 '{"name":"Alice","age":30}' EX 3600
GET cache:user:123
```

### File de tâches
```bash
RPUSH tasks "send_email" "process_image"
LPOP tasks  # Consomme une tâche
```

### Compteurs temps réel
```bash
INCR page:views
INCRBY user:123:score 10
```

### Objets structurés
```bash
HSET user:123 name "Alice" email "alice@example.com"
HGETALL user:123
```

---

## Roadmap

### Prochaines fonctionnalités (à voir ?)
- [ ] **Persistence**: RDB snapshots + AOF logs
- [ ] **Pub/Sub**: PUBLISH/SUBSCRIBE temps réel
- [ ] **Transactions**: MULTI/EXEC/WATCH
- [ ] **Sorted Sets**: ZADD/ZRANGE avec scores
- [ ] **Clustering**: Distribution horizontale