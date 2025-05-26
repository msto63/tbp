## Phase 1: Fundamentale Grundbausteine (Woche 1-2)

**1. Core-Modul**
- `core/context.go` - Erweiterte Context-Verwaltung mit User-ID, Tenant-ID, Request-ID
- `core/errors.go` - Basis-Fehlertypen und grundlegende Fehlerbehandlung
- `core/types.go` - Gemeinsame Typen und Interfaces
- `core/version.go` - Versionsinformationen

**2. Errors-Modul**
- `errors/errors.go` - Strukturierte Fehlerbehandlung mit Codes und Wrapping
- `errors/codes.go` - Fehlercodes für die Business-Logik
- `errors/details.go` - Fehlerdetails und Metadaten

## Phase 2: Konfiguration und Logging (Woche 2-3)

**3. Config-Modul**
- `config/config.go` - Basis-Konfigurationsmanagement
- `config/env.go` - Environment-Variable-Handling
- `config/file.go` - Dateibasierte Konfiguration

**4. Logging-Modul**
- `logging/logger.go` - Strukturiertes Logging
- `logging/context.go` - Context-basiertes Logging
- `logging/fields.go` - Strukturierte Felder

## Phase 3: Repository-Pattern für Taskmanagement (Woche 3-4)

**5. Patterns/Repository**
- `patterns/repository/repository.go` - Generische Repository-Interfaces
- `patterns/repository/memory.go` - In-Memory-Implementation für Tests
- `patterns/repository/sql.go` - SQL-Datenbank-Implementation

## Phase 4: Utilities für TCOL (Woche 4)

**6. StringX für TCOL-Parsing**
- `utils/stringx/abbrev.go` - Abkürzungserkennung für TCOL-Befehle
- `utils/stringx/similarity.go` - String-Ähnlichkeit für Fuzzy-Matching

## Phase 5: Testing-Framework (Woche 4-5)

**7. Testing-Utilities**
- `testing/mock.go` - Mocking-Utilities
- `testing/fixtures.go` - Test-Datenstrukturen
- `testing/time.go` - Time-Mocking für deterministische Tests

## Warum diese Reihenfolge?

1. **Core & Errors zuerst**: Ohne solide Grundlagen für Context und Fehlerbehandlung können andere Module nicht sinnvoll entwickelt werden.

2. **Config & Logging früh**: Diese werden von praktisch allen anderen Komponenten benötigt.

3. **Repository-Pattern für MVP**: Das Taskmanagement benötigt Datenpersistierung - das Repository-Pattern ist dafür essentiell.

4. **StringX für TCOL**: Die Abkürzungserkennung ist ein Kernfeature von TCOL und wird vom Application-Server benötigt.

5. **Testing parallel**: Testing-Utilities sollten parallel zu den anderen Modulen entwickelt werden, um von Anfang an hohe Codequalität sicherzustellen.

**Nicht benötigt für MVP:**
- Security-Module (können später hinzugefügt werden)
- Command-Pattern (kommt mit TCOL-Implementation)
- Events/Workflow (für spätere Erweiterungen)
- Metrics (Optimierung nach MVP)
