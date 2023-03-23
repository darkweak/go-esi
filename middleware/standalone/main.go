package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/darkweak/go-esi/esi"
)

const expected = `<html>
    <head>
        <title>Hello from domain.com:81</title>
        
    </head>
    <body>
        <esi:include src="domain.com/not-interpreted"/>
        <h1>CHAINED 2</h1>
        <h1>CHAINED 2</h1>
    </body>
</html>
`

func main() {
	buf := bytes.NewBuffer([]byte{})
	ctx := context.Background()
	rq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://domain.com:81/", nil)

	res, err := http.DefaultClient.Do(rq)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	_, _ = io.Copy(buf, res.Body)

	result := buf.String()
	parsed := esi.Parse(buf.Bytes(), rq)

	if string(parsed) != expected {
		fmt.Printf("Given:\n%+v\nParsed result:\n%+v\n", result, string(parsed))
		panic(nil)
	}
}
