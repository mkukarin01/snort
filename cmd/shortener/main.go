package main

import (
    "log"
    "net/http"

    "github.com/mkukarin01/snort/internal/app"
)

func main() {
    r := app.NewRouter()
    log.Fatal(http.ListenAndServe(":8080", r))
}
