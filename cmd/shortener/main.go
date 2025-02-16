package main

import (
    "log"
    "net/http"

    "github.com/mkukarin01/snort/internal/app"
)

func main() {
    srv := app.NewServer()
    log.Fatal(http.ListenAndServe(":8080", srv))
}
