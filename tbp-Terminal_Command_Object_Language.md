# TCOL - Terminal Command Object Language
## Fachliches und Technisches Konzept

### 1. Executive Summary

TCOL (Terminal Command Object Language) ist eine objektzentrierte Kommandosprache für ein TUI-basiertes Microservice-Anwendungssystem. Sie kombiniert die Effizienz klassischer Terminal-Kommandos mit modernen objektorientierten Paradigmen und ermöglicht intuitive, abkürzbare Befehle für Enterprise-Anwendungen.

**Kernmerkmale:**
- Vollständig objektbasierte Syntax
- Intelligente Abkürzungen bis zur Eindeutigkeit
- Konsistente Methoden-Notation
- Umfassendes Alias-System
- Multi-Service-Unterstützung

### 2. Architektur-Übersicht

```
┌─────────────────────┐
│   TUI Client        │
│  ┌───────────────┐  │
│  │ Input Parser  │  │
│  └───────┬───────┘  │
│          │          │
│  ┌───────▼───────┐  │
│  │ Alias Resolver│  │
│  └───────┬───────┘  │
│          │          │
│  ┌───────▼───────┐  │
│  │Command Builder│  │
│  └───────┬───────┘  │
└──────────┼──────────┘
           │ gRPC
┌──────────▼──────────┐
│ Application Server  │
│  ┌───────────────┐  │
│  │Command Parser │  │
│  └───────┬───────┘  │
│          │          │
│  ┌───────▼───────┐  │
│  │  Validator    │  │
│  └───────┬───────┘  │
│          │          │
│  ┌───────▼───────┐  │
│  │ Router/Exec   │  │
│  └───────┬───────┘  │
└──────────┼──────────┘
           │ gRPC
┌──────────▼──────────┐
│   Microservices     │
└─────────────────────┘
```

### 3. Syntax-Spezifikation

#### 3.1 Grammatik (EBNF)

```ebnf
command        = object_command | system_command | alias_command
object_command = object_spec "." method_name [parameters]
system_command = system_object "." method_name [parameters]
alias_command  = alias_name [parameters]

object_spec    = object_type "[" selector "]" | object_type ":" identifier
object_type    = IDENTIFIER
selector       = "*" | identifier | filter_expression
identifier     = ALPHANUMERIC+
filter_expression = condition {"," condition}
condition      = field_name operator value

method_name    = IDENTIFIER {"-" IDENTIFIER}
parameters     = parameter {"," parameter}
parameter      = param_name "=" param_value
param_value    = string | number | boolean | array

system_object  = "SYSTEM" | "QUERY" | "REPORT" | "BATCH" | "TRANSACTION" | 
                 "VARIABLE" | "MACRO" | "ALIAS" | "HELP"
```

#### 3.2 Abkürzungsregeln

**Algorithmus für Abkürzungserkennung:**

1. **Token-basierte Abkürzung**: Jeder Teil eines Befehls (getrennt durch "-") kann einzeln abgekürzt werden
   ```
   SHOW-SERVICE-STATUS → SH-SERV-ST → S-S-S
   ```

2. **Eindeutigkeitsprüfung**: 
   - Parser führt Präfix-Baum (Trie) für alle bekannten Kommandos
   - Abkürzung muss eindeutig auf genau ein Kommando zeigen
   - Bei Mehrdeutigkeit: Fehler mit Vorschlagsliste

3. **Kontextabhängige Auflösung**:
   - Objektmethoden werden im Kontext des Objekttyps aufgelöst
   - Systemmethoden global

### 4. Objektmodell

#### 4.1 Basis-Objekttypen

```
OBJECT
├── BUSINESS_OBJECT
│   ├── CUSTOMER
│   ├── INVOICE
│   ├── ORDER
│   └── PRODUCT
├── SYSTEM_OBJECT
│   ├── SYSTEM
│   ├── SERVICE
│   ├── USER
│   └── SESSION
└── UTILITY_OBJECT
    ├── QUERY
    ├── REPORT
    ├── BATCH
    └── TRANSACTION
```

#### 4.2 Objekt-Methoden-Matrix

| Objekttyp | Standard-Methoden | Spezial-Methoden |
|-----------|------------------|------------------|
| CUSTOMER | CREATE, SHOW, UPDATE, DELETE, LIST | ACTIVATE, DEACTIVATE, MERGE, VALIDATE |
| INVOICE | CREATE, SHOW, UPDATE, DELETE, LIST | SEND, MARK-PAID, CANCEL, ARCHIVE |
| QUERY | NEW, EXECUTE, SAVE, DELETE, LIST | SET-SOURCE, SET-FILTER, SET-FIELDS |
| SYSTEM | SHOW-STATUS, SHUTDOWN, RESTART | BACKUP, RESTORE, CONFIGURE |

### 5. Kommando-Verarbeitung

#### 5.1 Parse-Pipeline

```
1. Tokenization
   Input: "CUST:12345:email='new@example.com'"
   Tokens: ["CUST", ":", "12345", ":", "email", "=", "'new@example.com'"]

2. Alias Resolution
   Check: Ist "CUST" ein Alias?
   Result: "CUST" → "CUSTOMER"

3. Abbreviation Expansion
   Expand: "CUSTOMER" ist vollständig

4. Syntax Validation
   Check: Entspricht Grammatik für Kurzform-Update

5. Command Construction
   Build: CommandObject{
     Type: "CUSTOMER",
     ID: "12345",
     Method: "UPDATE",
     Params: {email: "new@example.com"}
   }

6. Security Check
   Verify: User has CUSTOMER:UPDATE permission

7. Routing
   Route: To Customer-Service
```

#### 5.2 Fehlerbehandlung

```
Fehlertypen:
- SYNTAX_ERROR: Ungültige Kommandosyntax
- AMBIGUOUS_COMMAND: Abkürzung nicht eindeutig
- OBJECT_NOT_FOUND: Objekt existiert nicht
- METHOD_NOT_FOUND: Methode für Objekt nicht verfügbar
- PERMISSION_DENIED: Keine Berechtigung
- SERVICE_UNAVAILABLE: Ziel-Service nicht erreichbar
- VALIDATION_ERROR: Parameter-Validierung fehlgeschlagen
```

### 6. Alias-System

#### 6.1 Alias-Speicherung

```sql
-- Alias-Tabelle
CREATE TABLE aliases (
    id UUID PRIMARY KEY,
    user_id VARCHAR(50),
    name VARCHAR(50),
    command TEXT,
    parameters TEXT[], -- Platzhalter-Definition
    scope VARCHAR(20), -- 'personal', 'team', 'global'
    context VARCHAR(50), -- Optional: Kontext-Einschränkung
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP,
    last_used TIMESTAMP
);

-- Alias-Permissions
CREATE TABLE alias_permissions (
    alias_id UUID,
    required_role VARCHAR(50),
    FOREIGN KEY (alias_id) REFERENCES aliases(id)
);
```

#### 6.2 Alias-Auflösung

```go
type AliasResolver struct {
    userAliases   map[string]*Alias
    teamAliases   map[string]*Alias
    globalAliases map[string]*Alias
}

func (r *AliasResolver) Resolve(name string, context string) (*Alias, error) {
    // Priorität: user > team > global
    // Berücksichtigt Kontext-Filter
}
```

### 7. Sicherheitskonzept

#### 7.1 Berechtigungsmodell

```
Permission := Object + Method + Scope

Beispiele:
- CUSTOMER:*:own          (Nur eigene Kunden)
- CUSTOMER:READ:all       (Alle Kunden lesen)
- CUSTOMER:DELETE:none    (Kein Löschen)
- INVOICE:CREATE:limited  (Limitiert auf Betrag)
```

#### 7.2 Audit-Trail

```go
type AuditEntry struct {
    Timestamp   time.Time
    UserID      string
    SessionID   string
    Command     string
    ObjectType  string
    ObjectID    string
    Method      string
    Parameters  map[string]interface{}
    Result      string
    Duration    time.Duration
    ErrorCode   string
}
```

### 8. Performance-Optimierung

#### 8.1 Command-Cache

```go
type CommandCache struct {
    // LRU-Cache für geparste Kommandos
    cache *lru.Cache
    
    // Vorcompilierte Patterns
    patterns map[string]*regexp.Regexp
    
    // Trie für Abkürzungen
    abbreviationTrie *Trie
}
```

#### 8.2 Batch-Optimierung

- Kommando-Batching für Bulk-Operationen
- Parallele Ausführung wo möglich
- Transaktionale Gruppierung

### 9. Erweiterbarkeit

#### 9.1 Service-Integration

Neue Services registrieren ihre Objekte und Methoden:

```protobuf
message ServiceRegistration {
    string service_id = 1;
    repeated ObjectDefinition objects = 2;
}

message ObjectDefinition {
    string object_type = 1;
    repeated MethodDefinition methods = 2;
    repeated string selectable_fields = 3;
}

message MethodDefinition {
    string name = 1;
    repeated ParameterDefinition parameters = 2;
    repeated string required_permissions = 3;
}
```

#### 9.2 Plugin-System

```go
type CommandPlugin interface {
    // Plugin-Metadaten
    GetMetadata() PluginMetadata
    
    // Kommando-Erweiterungen
    GetCommands() []CommandDefinition
    
    // Ausführung
    Execute(ctx context.Context, cmd Command) (Result, error)
}
```

### 10. Implementierungsroadmap

**Phase 1: Core Implementation (4-6 Wochen)**
- Basis-Parser und Grammatik
- Grundlegende Objekttypen (SYSTEM, CUSTOMER, INVOICE)
- Einfache Methoden (CREATE, SHOW, UPDATE, DELETE, LIST)
- Basis-Fehlerbehandlung

**Phase 2: Advanced Features (4-6 Wochen)**
- Vollständiges Abkürzungssystem
- Alias-System
- Batch-Verarbeitung
- Transaktionen

**Phase 3: Enterprise Features (6-8 Wochen)**
- Komplexe Selektoren und Filter
- Query-Objekt mit SQL-ähnlicher Syntax
- Report-System
- Audit-Trail und Sicherheit

**Phase 4: Optimization & Polish (4 Wochen)**
- Performance-Optimierung
- Erweiterte Fehlerbehandlung
- Dokumentation
- Tooling (Syntax-Highlighting, Auto-Completion)

### 11. Beispiel-Kommandos (Zusammenfassung)

```bash
# Basis-Operationen
CUSTOMER.CREATE name="Example Corp" type="B2B"
CUST:12345                                      # Kurzform für SHOW
CUSTOMER[city="Berlin"].LIST
CUSTOMER:12345:email="new@example.com"          # Kurzform für UPDATE

# Komplexe Operationen
QUERY.EXECUTE source="INVOICE" filter="unpaid=true AND age>30"
BATCH.EXECUTE file="month-end.tcl" mode="transaction"
REPORT[monthly-sales].GENERATE format="pdf"

# Alias-Definitionen
ALIAS.CREATE name="uc" command="CUSTOMER.LIST filter='unpaid=true'"
ALIAS.CREATE name="morning" command=["SYSTEM.STATUS", "ORDER.LIST filter='today'"]

# System-Operationen
SYSTEM.SHOW-STATUS
SERVICE[inventory].RESTART
TRANSACTION.EXECUTE commands=["ACCOUNT[1].WITHDRAW 1000", "ACCOUNT[2].DEPOSIT 1000"]
```

Dieses Konzept bietet eine solide Grundlage für die Implementierung von TCOL und kann je nach Anforderungen erweitert werden.

