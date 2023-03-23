package esi_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/darkweak/go-esi/esi"
)

func loadFromFixtures(name string) []byte {
	b, e := os.ReadFile("../fixtures/" + name)
	if e != nil {
		panic("The file " + name + " doesn't exist.")
	}

	return b
}

func getRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "http://domain.com:9080", nil)
}

var expected = map[string]string{
	"include": `<h1>CHAINED 2</h1>`,
	"comment": `<h1>CHAINED 2</h1>`,
	"choose": `
        <div>
            <h1>CHAINED 2</h1>
        </div>
    `,
	"full.html": `<html>
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
`,
	"escape": `<p><esi:vars>Hello, $(HTTP_COOKIE{name})!</esi:vars></p>`,
	"remove": `<h1>CHAINED 2</h1>

<h1>CHAINED 2</h1>
<h1>CHAINED 2</h1>`,
	"vars": `
  <img src="http://www.example.com/my_value/hello.gif" />
  <img src="http://www.example.com/default_value/hello.gif" />
  <img src="http://www.example.com/domain.com:9080/hello.gif"/>
  <img src="http://www.example.com/true/hello.gif"/ >`,
}

func verify(t *testing.T, fixture string) {
	t.Helper()

	if result := string(esi.Parse(loadFromFixtures(fixture), getRequest())); result != expected[fixture] {
		t.Errorf("ESI parsing mismatch from `%s` expected\nExpected:\n%+v\nGiven:\n%+v\n", fixture, expected[fixture], result)
	}
}

func Test_Parse_includeMock(t *testing.T) {
	t.Parallel()
	verify(t, "include")
}

func Test_Parse_commentMock(t *testing.T) {
	t.Parallel()
	verify(t, "comment")
}

func Test_Parse_chooseMock(t *testing.T) {
	t.Parallel()
	verify(t, "choose")
}

func Test_Parse_escapeMock(t *testing.T) {
	t.Parallel()
	verify(t, "escape")
}

func Test_Parse_removeMock(t *testing.T) {
	t.Parallel()
	verify(t, "remove")
}

func Test_Parse_varsMock(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://domain.com:9080", nil)
	req.AddCookie(&http.Cookie{
		Name:  "type",
		Value: "my_value",
	})
	req.Header.Add("Accept-Language", "en")

	if result := string(esi.Parse(loadFromFixtures("vars"), req)); result != expected["vars"] {
		t.Errorf("ESI parsing mismatch from `%s` expected\nExpected:\n%+v\nGiven:\n%+v\n", "vars", expected["vars"], result)
	}
}

func Test_Parse_fullMock(t *testing.T) {
	t.Parallel()
	verify(t, "full.html")
}

// Benchmarks.
func BenchmarkInclude(b *testing.B) {
	for i := 0; i < b.N; i++ {
		esi.Parse(
			[]byte(
				`<esi:include src="http://domain.com:9080/chained-esi-include-1" alt=http://domain.com:9080/alt-esi-include/>`,
			),
			httptest.NewRequest(http.MethodGet, "http://domain.com:9080", nil),
		)
	}
}

var remove = `<esi:include src="http://domain.com:9080/chained-esi-include-1"/>
<esi:remove>
  <a href="http://www.example.com">www.example.com</a>
</esi:remove>
<esi:include src="http://domain.com:9080/chained-esi-include-1"/>
<esi:include src="http://domain.com:9080/chained-esi-include-1"/>`

func BenchmarkRemove(b *testing.B) {
	for i := 0; i < b.N; i++ {
		esi.Parse(
			[]byte(remove),
			httptest.NewRequest(http.MethodGet, "http://domain.com:9080", nil),
		)
	}
}

const full = `<html>
<head>
	<title><esi:vars>Hello from $(HTTP_HOST)</esi:vars></title>
	<esi:remove>
		<esi:include src="http://domain.com:9080/chained-esi-include-1" />
	</esi:remove>
</head>
<body>
	<!--esi
	<esi:include src="domain.com:9080/not-interpreted"/>
	-->
	<esi:include src="/chained-esi-include-1" />
	<esi:include src="http://inexistent.abc/something" alt="//domain.com:9080/alt-esi-include" onerror="continue" />
	<esi:choose> 
		<esi:when test="$(HTTP_COOKIE{group})=='Advanced'"> 
			<span><esi:include src="http://domain.com:9080/chained-esi-include-1"/></span>
		</esi:when> 
		<esi:when test="$(HTTP_COOKIE{group})=='Basic User'">
			<esi:include src="https://google.com"/>
		</esi:when> 
		<esi:otherwise> 
			<div>
				<esi:include src="http://domain.com:9080/esi-include"/>
			</div>
		</esi:otherwise>
	</esi:choose>
</body>
</html>
`

func BenchmarkFull(b *testing.B) {
	for i := 0; i < b.N; i++ {
		esi.Parse(
			[]byte(full),
			httptest.NewRequest(http.MethodGet, "http://domain.com:9080", nil),
		)
	}
}
