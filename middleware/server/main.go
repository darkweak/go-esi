package main

import (
	"net/http"

	"github.com/darkweak/go-esi/esi"
)

func main() {
	rq, _ := http.NewRequest(http.MethodGet, "domain.com/", nil)
	esi.Parse([]byte{}, rq)
}
