package caddy_esi

import (
	"net/http"
	"testing"

	"github.com/caddyserver/caddy/v2/caddytest"
)

const expectedOutput = `<html>
    <head>
        <title>Hello from domain.com:9080</title>
        
    </head>
    <body>
        <esi:include src="domain.com:9080/not-interpreted"/>
        <h1>CHAINED 2</h1>
        <h1>ALTERNATE ESI INCLUDE</h1>
         
                <div>
                    <h1>ESI INCLUDE</h1>
                </div>
            
    </body>
</html>
`

func loadConfiguration(t *testing.T) *caddytest.Tester {
	tester := caddytest.NewTester(t)
	tester.InitServer(`
	{
		admin localhost:2999
		order esi before basicauth
		esi
		http_port 9080
	}
	domain.com:9080 {
		route /chained-esi-include-1 {
			header Content-Type text/html
			respond `+"`<esi:include src=\"/chained-esi-include-2\"/>`"+`
		}
	
		route /chained-esi-include-2 {
			header Content-Type text/html
			respond "<h1>CHAINED 2</h1>"
		}
	
		route /esi-include {
			header Content-Type text/html
			respond "<h1>ESI INCLUDE</h1>"
		}
	
		route /alt-esi-include {
			header Content-Type text/html
			respond "<h1>ALTERNATE ESI INCLUDE</h1>"
		}
	
		route /* {
			esi
			root * ../../fixtures
			file_server
		}
	}`, "caddyfile")

	return tester
}

func TestFullHTML(t *testing.T) {
	tester := loadConfiguration(t)
	_, _ = tester.AssertGetResponse(`http://domain.com:9080/full.html`, http.StatusOK, expectedOutput)
}

func TestUnitary(t *testing.T) {
	tester := loadConfiguration(t)

	_, _ = tester.AssertGetResponse(`http://domain.com:9080/escape`, http.StatusOK, `<p><esi:vars>Hello, $(HTTP_COOKIE{name})!</esi:vars></p>`)
	_, _ = tester.AssertGetResponse(`http://domain.com:9080/include`, http.StatusOK, "<h1>CHAINED 2</h1>")
	_, _ = tester.AssertGetResponse(`http://domain.com:9080/alt`, http.StatusOK, "<h1>ALTERNATE ESI INCLUDE</h1>")
}

func TestDefaultHost(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(`
	{
		admin localhost:2999
		order esi before basicauth
		esi
		http_port 9080
	}
	localhost:9080 {
		respond /my-esi-content/whatever "<h1>FROM DEFAULT HOST</h1>"
	}
	domain.com:9080 {
		route /page {
			header Content-Type text/html
			esi {
				default_host localhost:9080
			}
			respond `+"`<esi:include src=\"/my-esi-content/whatever\"/>`"+`
		}
	}`, "caddyfile")

	rq, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:9080/page", nil)
	rq.Host = "domain.com"

	_, _ = tester.AssertResponse(rq, http.StatusOK, "<h1>FROM DEFAULT HOST</h1>")
}
