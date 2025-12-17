package main

import (
    "fmt"
    "log"
    "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    fmt.Fprint(w, "Hello from Agentry")
}

func main() {
    http.HandleFunc("/", helloHandler)
    addr := ":8090"
    log.Printf("Listening on %s...", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
