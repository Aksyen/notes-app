package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB // ← ГЛОБАЛЬНАЯ переменная

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	formatted := time.Time(t).UTC().Format("2006-01-02T15:04:05Z")
	return []byte(`"` + formatted + `"`), nil
}

type Note struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	CreatedAt JSONTime `json:"created_at"`
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func getNotesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content, created_at FROM notes ORDER BY id")
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}
		notes = append(notes, n)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func postNotesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var note Note
	err := db.QueryRow(`
		INSERT INTO notes (title, content)
		VALUES ($1, $2)
		RETURNING id, title, content, created_at
	`, input.Title, input.Content).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
	if err != nil {
		http.Error(w, "DB insert failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func notesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getNotesHandler(w, r)
	case "POST":
		postNotesHandler(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=appuser password=secret dbname=notesdb sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping DB:", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	log.Println("✅ Table 'notes' ready")

	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/notes", notesHandler)

	log.Println("🚀 Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	//	curl -X POST localhost:8080/notes -H "Content-Type: application/json" -d '{"title":"Из БД","content":"Ура!"}'
	//
	// curl localhost:8080/notes
}
