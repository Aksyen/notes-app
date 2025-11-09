package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	// Форматируем время как RFC3339, но без наносекунд
	formatted := time.Time(t).UTC().Format("2006-01-02T15:04:05Z")
	return []byte(`"` + formatted + `"`), nil
}

// Note — структура одной заметки
type Note struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	CreatedAt JSONTime `json:"created_at"`
}

// Глобальная переменная — наше временное "хранилище"
var notes []Note
var nextID = 1 // чтобы ID росли: 1, 2, 3...

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

// Новый хендлер: отдаёт все заметки
func getNotesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes) // просто отдаём весь срез
}

func main() {
	// Добавим одну заметку "вручную", чтобы было что посмотреть
	notes = append(notes, Note{
		ID:        nextID,
		Title:     "Добро пожаловать!",
		Content:   "Это первая заметка в нашем API.",
		CreatedAt: JSONTime(time.Now()),
	})
	nextID++

	// Регистрируем новый путь
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/notes", getNotesHandler) // ← новый!

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
