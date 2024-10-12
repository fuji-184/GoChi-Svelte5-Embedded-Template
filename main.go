package main

import (
    "embed"
    "fmt"
    "io/fs"
    "log"
    "net/http"
    "github.com/go-chi/chi/v5"
)

//go:embed all:svelte/build
var svelteStatic embed.FS

func main() {

    s, err := fs.Sub(svelteStatic, "svelte/build")
    if err != nil {
        panic(err)
    }

    staticServer := http.FileServer(http.FS(s))

    r := chi.NewRouter()

    r.Handle("/", staticServer)
    r.Handle("/_app/*", staticServer)
	r.Handle("/service-worker.js", staticServer)
    r.Handle("/favicon.png", staticServer)
    r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
        r.URL.Path = "/"
        staticServer.ServeHTTP(w, r)
    })

    fmt.Println("Running on port: 3000")
    log.Fatal(http.ListenAndServe(":3000", r))
}