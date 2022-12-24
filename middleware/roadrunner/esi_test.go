package roadrunner_esi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type (
	configWrapper struct{}
	next          struct{}
)

var nextFilter = &next{}

func (n *next) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	rw.WriteHeader(http.StatusOK)
	if rq.RequestURI == "/esi-include-1" {
		_, _ = rw.Write([]byte("Awesome first included ESI tag!"))

		return
	}
	if rq.RequestURI == "/esi-include-2" {
		_, _ = rw.Write([]byte("Another included ESI tag!"))

		return
	}
	_, _ = rw.Write([]byte(`Hello Roadrunner! <esi:include src="/include" />`))
}

func (*configWrapper) Get(name string) any {
	return nil
}
func (*configWrapper) Has(name string) bool {
	return true
}

func Test_Plugin_Init(t *testing.T) {
	p := &Plugin{}

	if p.Init(&configWrapper{}, nil) != nil {
		t.Error("The Init method must not crash if a valid configuration is given.")
	}

	defer func() {
		if recover() == nil {
			t.Error("The Init method must crash if a nil configuration is given.")
		}
	}()
	err := p.Init(nil, nil)
	if err != nil {
		t.Error(err.Error())
	}
}

func prepare(endpoint string) (req *http.Request, res1 *httptest.ResponseRecorder) {
	req = httptest.NewRequest(http.MethodGet, endpoint, nil)
	res1 = httptest.NewRecorder()

	return
}

func Test_Plugin_Middleware(t *testing.T) {
	p := &Plugin{}
	_ = p.Init(&configWrapper{}, nil)
	handler := p.Middleware(nextFilter)
	req, res := prepare("http://localhost/handled")
	handler.ServeHTTP(res, req)
	rs := res.Result()
	defer rs.Body.Close()
	b, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Error("body read error")
	}

	if string(b) != "Hello Roadrunner! " {
		t.Error(`The returned response must be equal to "Hello Roadrunner! " because of non running service.`)
	}
}
