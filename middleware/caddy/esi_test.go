package caddy_esi

import (
	"net/http"
	"os"
	"testing"

	"github.com/caddyserver/caddy/v2/caddytest"
)

const expectedOutput = `<html>
    <head>
        <title>Hello from domain.com:9080</title>
        
    </head>
    <body>
        
        <esi:include src="domain.com/not-interpreted"/>
        
        <h1>CHAINED 2</h1>
        <h1>ALTERNATE ESI INCLUDE</h1>
         
                <div>
                    
                </div>
            
    </body>
</html>
`

func loadCaddyfile() string {
	b, _ := os.ReadFile("./Caddyfile")
	return string(b)
}

func TestFullHTML(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(loadCaddyfile(), "caddyfile")

	_, _ = tester.AssertGetResponse(`http://domain.com:9080/full.html`, http.StatusOK, expectedOutput)
}

func TestInclude(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(loadCaddyfile(), "caddyfile")

	_, _ = tester.AssertGetResponse(`http://domain.com:9080/include`, http.StatusOK, "<h1>CHAINED 2</h1>")
}

func TestIncludeAlt(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(loadCaddyfile(), "caddyfile")

	_, _ = tester.AssertGetResponse(`http://domain.com:9080/alt`, http.StatusOK, "<h1>ALTERNATE ESI INCLUDE</h1>")
}

func TestEscape(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(loadCaddyfile(), "caddyfile")

	_, _ = tester.AssertGetResponse(`http://domain.com:9080/escape`, http.StatusOK, `  
  <p><esi:vars>Hello, $(HTTP_COOKIE{name})!</esi:vars></p>
`)
}
