# TCOL Integration mit n8n

## Architektur-Konzept

```
┌─────────────────────┐         ┌─────────────────────┐
│   TUI Client        │         │   n8n Instance      │
│                     │         │                     │
│  WORKFLOW.START     │         │  - Workflow Engine  │
│  WORKFLOW.STATUS    │         │  - REST API         │
└──────────┬──────────┘         │  - Webhooks        │
           │                    └──────────▲──────────┘
           │ gRPC                          │ RESTWebhook
┌──────────▼──────────┐         ┌─────────┴──────────┐
│ Application Server  │────────▶│  n8n Service       │
│                     │  gRPC   │  Adapter           │
│  - Command Router   │         │                    │
│  - Service Registry │         │  - n8n API Client  │
│                     │         │  - Status Tracker  │
└─────────────────────┘         │  - Event Handler   │
                                └────────────────────┘
```

## WORKFLOW Objekt-Definition

```bash
# Basis-Objekt für n8n-Integration
WORKFLOW
├── Methods
│   ├── LIST              # Verfügbare Workflows
│   ├── START             # Workflow starten
│   ├── STOP              # Laufenden Workflow stoppen
│   ├── STATUS            # Status abfragen
│   ├── SHOW              # Details anzeigen
│   ├── HISTORY           # Ausführungshistorie
│   ├── CREATE            # Neuen Workflow erstellen
│   ├── UPDATE            # Workflow aktualisieren
│   ├── DELETE            # Workflow löschen
│   ├── IMPORT            # Aus JSON importieren
│   └── EXPORT            # Als JSON exportieren
```

## Kommando-Beispiele

### 1. Workflow-Verwaltung

```bash
# Alle verfügbaren Workflows anzeigen
WORKFLOW.LIST
WORKFLOW.LIST filter=tag=finance
WORKFLOW.LIST filter=active=true

# Workflow-Details
WORKFLOW[invoice-processing].SHOW
WORKFLOW[customer-onboarding].SHOW-NODES

# Workflow aktivierendeaktivieren
WORKFLOW[daily-report].ACTIVATE
WORKFLOW[old-process].DEACTIVATE
```

### 2. Workflow-Ausführung

```bash
# Einfacher Start
WORKFLOW[invoice-processing].START

# Mit Parametern
WORKFLOW[invoice-processing].START 
  invoice_id=INV-2024-001 
  customer=ACME-CORP 
  amount=5000

# Mit Kontext aus TCOL-Objekten
INVOICE[2024-001].TRIGGER-WORKFLOW name=invoice-processing
# Übergibt automatisch alle Invoice-Daten an n8n

# Batch-Verarbeitung
INVOICE[status=new].TRIGGER-WORKFLOW 
  name=invoice-processing 
  mode=batch 
  limit=50
```

### 3. Integration mit Zeiterfassung Use-Case

```bash
# Workflow für Monatsabschluss
WORKFLOW[timesheet-monthly-close].START 
  month=2024-01 
  department=IT

# Automatische Zeiterfassungs-Validierung
TRIGGER.CREATE name=validate-timesheet
  ON=TIMESHEET.SUBMIT
  ACTION=WORKFLOW[timesheet-validation].START employee=$user week=$week

# Überstunden-Benachrichtigung
WORKFLOW[overtime-notification].START 
  threshold=10 
  period=current-week
```

### 4. Status und Monitoring

```bash
# Laufende Workflows
WORKFLOW.STATUS filter=running
WORKFLOW.STATUS filter=failed AND date=today

# Spezifischer Workflow-Status
WORKFLOW[invoice-processing].STATUS execution=12345
WORKFLOW[invoice-processing].SHOW-LOG execution=12345

# Execution History
WORKFLOW[daily-report].HISTORY limit=10
WORKFLOW[daily-report].HISTORY filter=status=error

# Metriken
WORKFLOW[invoice-processing].STATS period=last-month
# Zeigt Ausführungen, Erfolgsrate, Durchschnittsdauer, etc.
```

### 5. Workflow-Erstellung via TCOL

```bash
# Neuen Workflow erstellen
WORKFLOW.CREATE name=customer-followup 
  description=Automatische Kundenbetreuung 
  trigger=schedule 
  schedule=0 9   MON

# Nodes hinzufügen
WORKFLOW[customer-followup].ADD-NODE 
  type=tcol 
  name=get-customers 
  command=CUSTOMER.LIST filter='last_contact30days'

WORKFLOW[customer-followup].ADD-NODE 
  type=email 
  name=send-email 
  template=followup

WORKFLOW[customer-followup].CONNECT 
  from=get-customers 
  to=send-email
```

### 6. TCOL-Node für n8n

```bash
# n8n kann TCOL-Kommandos ausführen
# Spezieller Node-Type TCOL Execute

# Beispiel n8n Workflow Node
{
  type tcol-execute,
  parameters {
    command INVOICE.CREATE,
    arguments {
      customer {{$node.customer.id}},
      amount {{$node.calculation.total}},
      items {{$node.items.json}}
    }
  }
}

# TCOL-Trigger für n8n
WORKFLOW.REGISTER-TRIGGER 
  name=new-customer 
  event=CUSTOMER.CREATE 
  workflow=customer-onboarding
```

### 7. Praktische Aliase

```bash
# Workflow-Aliase
ALIAS.CREATE name=wf command=WORKFLOW.LIST
ALIAS.CREATE name=wf-run command=WORKFLOW[$1].START
ALIAS.CREATE name=wf-status command=WORKFLOW.STATUS filter='running'

# Spezifische Business-Workflows
ALIAS.CREATE name=monatsabschluss command=
  WORKFLOW[monthly-closing].START month='$1'
  WORKFLOW[monthly-closing].STATUS follow=true


# Fehlerbehandlung
ALIAS.CREATE name=wf-errors command=WORKFLOW.STATUS filter='failed AND date=today'  SHOW-DETAILS
```

### 8. Event-Driven Integration

```bash
# TCOL Events triggern n8n Workflows
EVENT.MAP source=INVOICE.PAID target=WORKFLOW[invoice-paid-process].START
EVENT.MAP source=CUSTOMER.CHURN-RISK target=WORKFLOW[retention-campaign].START

# n8n Webhooks registrieren
WEBHOOK.CREATE 
  name=n8n-callback 
  url=httpsapp-serverwebhookn8n 
  events=[workflow.completed, workflow.failed]

# Bidirektionale Kommunikation
WORKFLOW[data-sync].CALLBACK 
  on=completion 
  execute=REPORT.GENERATE type='sync-summary'
```

### 9. Error Handling & Recovery

```bash
# Fehlerbehandlung
WORKFLOW[invoice-processing].RETRY execution=12345
WORKFLOW[invoice-processing].RETRY-FAILED date=today

# Manueller Eingriff
WORKFLOW[stuck-process].EXECUTION[12345].RESUME
WORKFLOW[stuck-process].EXECUTION[12345].SKIP-NODE node=validation
WORKFLOW[stuck-process].EXECUTION[12345].CANCEL

# Automatische Recovery
TRIGGER.CREATE name=workflow-recovery
  ON=WORKFLOW.FAILED
  CONDITION=workflow.critical=true
  ACTION=NOTIFICATION.SEND to='ops-team' AND WORKFLOW.RETRY after='5min'
```

### 10. Workflow Templates & Marketplace

```bash
# Workflow-Templates durchsuchen
WORKFLOW.SEARCH-TEMPLATES keyword=invoice
WORKFLOW.INSTALL-TEMPLATE name=stripe-invoice-sync

# Eigene Templates erstellen
WORKFLOW[my-process].EXPORT-TEMPLATE 
  name=timesheet-approval 
  category=HR 
  share=team

# Workflow-Bausteine
WORKFLOW.SNIPPET.CREATE name=tcol-customer-check 
  description=Prüft Kundenstatus via TCOL 
  nodes=[...] 
  reusable=true
```

## Implementierungsdetails

### n8n Service Adapter

```go
type N8nService struct {
    apiClient     n8n.Client
    webhookServer WebhookServer
    eventBus      EventBus
    workflowCache map[string]Workflow
}

 Registriert TCOL-Commands als n8n-fähig
func (s N8nService) RegisterCapabilities() []Capability {
    return []Capability{
        {Object WORKFLOW, Methods []string{START, STOP, STATUS, ...}},
        {Object , Methods []string{TRIGGER-WORKFLOW}},
    }
}

 TCOL Custom Node für n8n
type TCOLNode struct {
    Command    string
    Parameters map[string]interface{}
    Timeout    time.Duration
}
```

### Vorteile der Integration

1. Automation Komplexe Geschäftsprozesse automatisieren
2. Flexibilität n8n's visuelle Workflows + TCOL's Kommando-Power
3. Integration Verbindung zu 200+ externen Services via n8n
4. Monitoring Zentrale Überwachung aller Automatisierungen
5. Self-Service Fachbereich kann eigene Workflows erstellen

Diese Integration würde TCOL zu einer vollständigen Business-Automation-Plattform machen!