package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type CommandRequest struct {
    Command string `json:"command"`
}

func main() {
    startBlogSync()

    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./static/index.html")
    })

    http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./static/index.html")
    })

    http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
        var req CommandRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid request", http.StatusBadRequest)
            return
        }
        output := runCommand(req.Command)
        w.Write([]byte(output))
    })

    fmt.Println("Server starting on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}