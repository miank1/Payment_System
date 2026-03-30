HTTP Request
     │
     ▼
┌─────────────┐
│   Handler   │  ← Receives HTTP request, validates JSON
│ (handler.go)│    Talks to: Service
└─────────────┘
     │
     ▼
┌─────────────┐
│   Service   │  ← Business logic lives here
│ (service.go)│    Generates ID, sets status, stamps time
└─────────────┘    Talks to: Repository
     │
     ▼
┌─────────────┐
│ Repository  │  ← Only place that talks to DB
│(repository  │    Runs SQL queries
│    .go)     │    Talks to: Postgres
└─────────────┘
     │
     ▼
┌─────────────┐
│  Postgres   │  ← Stores data
│  (Neon DB)  │
└─────────────┘