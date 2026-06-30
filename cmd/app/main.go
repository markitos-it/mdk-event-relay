package main

import (
	"context"
	"database/sql"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"markitos-it-event-relay/internal/clients"
	"markitos-it-event-relay/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	ID      int
	Event   string
	Payload string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("[INFO] Opening SQLite database...")
	db, err := sql.Open("sqlite3", "./events.db?_journal_mode=WAL")
	if err != nil {
		log.Fatalf("[FATAL] Error opening DB: %v", err)
	}
	defer db.Close()

	schema := `
    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		event TEXT NOT NULL,
        payload TEXT NOT NULL, 
        status TEXT NOT NULL DEFAULT 'pending'
    );`

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("[FATAL] Error creating database schema: %v", err)
	}
	log.Println("[INFO] Database schema verified and ready.")
	// ---------------------------------------------------------------------

	busType := os.Getenv("BUS_TYPE")
	projectID := os.Getenv("GCP_PROJECT_ID")
	topicID := os.Getenv("GCP_TOPIC_ID")

	var bus domain.EventPublisher

	switch busType {
	case "pubsub":
		bus, err = clients.NewPubSubPublisher(ctx, projectID, topicID)
	case "mock":
		bus = clients.NewConsolePublisher()
	default:
		log.Fatalf("[FATAL] Unknown BUS_TYPE: %s", busType)
	}

	if err != nil {
		log.Fatalf("[FATAL] Error initializing bus (%s): %v", busType, err)
	}

	log.Printf("[INFO] Started with mode: %s", busType)

	runRelayer(ctx, db, bus)
}

func runRelayer(ctx context.Context, db *sql.DB, bus domain.EventPublisher) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	backoffFactor := 0.0
	maxBackoff := 60.0

	log.Println("[INFO] Orchestrator active. Waiting for events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] Stop signal received. Diplomatic Corps is withdrawing.")
			return
		case <-ticker.C:
			rows, err := db.Query("SELECT id, event, payload FROM events WHERE status = 'pending' LIMIT 10")
			if err != nil {
				log.Printf("[ERROR] Error querying DB: %v", err)
				continue
			}

			var events []Event
			for rows.Next() {
				var e Event
				if err := rows.Scan(&e.ID, &e.Event, &e.Payload); err != nil {
					log.Printf("[ERROR] Error scanning event: %v", err)
					continue
				}
				events = append(events, e)
			}
			rows.Close()

			if len(events) == 0 {
				continue
			}

			for _, event := range events {
				log.Printf("[PROCESS] Delegating event %d to the Bus", event.ID)

				if err := bus.Publish(ctx, []byte(event.Payload)); err != nil {
					log.Printf("[WARN] The Ambassador failed: %v", err)

					wait := math.Min(math.Pow(2, backoffFactor), maxBackoff)
					log.Printf("[DEBUG] Waiting %v seconds", wait)

					time.Sleep(time.Duration(wait) * time.Second)
					backoffFactor = math.Min(backoffFactor+1, 6)
					break
				}

				_, err := db.Exec("UPDATE events SET status = 'sent' WHERE id = ?", event.ID)
				if err != nil {
					log.Printf("[ERROR] Failed to mark event %d: %v", event.ID, err)
				} else {
					log.Printf("[SUCCESS] Event %d delivered", event.ID)
					backoffFactor = 0
				}
			}
		}
	}
}
