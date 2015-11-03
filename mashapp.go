package main

import (
    "github.com/padster/go-sound/mashapp"
)

func main() {
    mashapp.NewServer(8080, "mashapp").Serve()
}
