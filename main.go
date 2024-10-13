package main

import (
    "database/sql"
    "embed"
    "encoding/json"
    "fmt"
    "io/fs"
    "log"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    _ "github.com/mattn/go-sqlite3"
)

//go:embed all:svelte/build
var svelteStatic embed.FS

type Tes struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    s, err := fs.Sub(svelteStatic, "svelte/build")
    if err != nil {
        panic(err)
    }

    staticServer := http.FileServer(http.FS(s))

    db, err := sql.Open("sqlite3", "./fuji.db")
    if err != nil {
        log.Fatal(err)
    }

    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(time.Hour)

    defer db.Close()

    r := chi.NewRouter()
    r.Handle("/", staticServer)
    r.Handle("/_app/*", staticServer)
    r.Handle("/service-worker.js", staticServer)
    r.Handle("/favicon.png", staticServer)

    r.Get("/tes", func(w http.ResponseWriter, r *http.Request) {
        handleTes(w, r, db)
    })

    r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
        r.URL.Path = "/"
        staticServer.ServeHTTP(w, r)
    })

    fmt.Println("Running on port: 8000")
    log.Fatal(http.ListenAndServe(":8000", r))
}

func handleTes(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS tes (id INTEGER NOT NULL PRIMARY KEY, name TEXT)`)
    if err != nil {
        http.Error(w, "Error creating table", http.StatusInternalServerError)
        return
    }

    stmt, err := db.Prepare("INSERT INTO tes(name) VALUES(?)")
    if err != nil {
        http.Error(w, "Error preparing insert statement", http.StatusInternalServerError)
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec("fuji")
    if err != nil {
        http.Error(w, "Error inserting data", http.StatusInternalServerError)
        return
    }

    rows, err := db.Query("SELECT id, name FROM tes")
    if err != nil {
        http.Error(w, "Error querying data", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var results []Tes
    for rows.Next() {
        var tes Tes
        if err := rows.Scan(&tes.Id, &tes.Name); err != nil {
            http.Error(w, "Error scanning row", http.StatusInternalServerError)
            return
        }
        results = append(results, tes)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}
