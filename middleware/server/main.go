package main

import (
	"context"
	"net/http"
	"time"

	"github.com/darkweak/go-esi/esi"
)

var respond = []byte(`<html>
<head>
	<title><esi:vars>Hello from $(HTTP_HOST)</esi:vars></title>
	<esi:remove>
		<esi:include src="http://domain.com/chained-esi-include-1" />
	</esi:remove>
</head>
<body>
	<!--esi
	<esi:include src="domain.com/not-interpreted"/>
	-->
	<esi:include src="http://domain.com/chained-esi-include-1" />
	<esi:include src="http://domain.com/chained-esi-include-1" />
</body>
</html>
`)

func main() {
	rq, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "domain.com/", nil)
	esi.Parse([]byte{}, rq)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respond[0:97])
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		time.Sleep(time.Second)
		_, _ = w.Write(respond[97:194])
		time.Sleep(time.Second)
		_, _ = w.Write(respond[194:291])
		time.Sleep(time.Second)
		_, _ = w.Write(respond[291:388])
	})

	server := &http.Server{
		Addr:              ":81",
		ReadHeaderTimeout: 3 * time.Second,
	}
	_ = server.ListenAndServe()
}
