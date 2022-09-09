package esi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	unaryNegationTest = []byte("!(1==1)")
	comparisonTest    = []byte("!('a'<='c')")
	logicalOrTest     = []byte("(1==1)|('abc'=='def')")
	logicalAndTest    = []byte("(4!=5)&(4==5)")
	complexTest       = []byte("$(HTTP_ACCEPT_LANGUAGE{en-gb})")
)

func Test_validateTest(t *testing.T) {
	t.Parallel()

	if validateTest(unaryNegationTest, httptest.NewRequest(http.MethodGet, "http://domain.com", nil)) {
		t.Error("The unaryNegationTest must return false because we return the opposite of true (1==1)")
	}

	if validateTest(comparisonTest, httptest.NewRequest(http.MethodGet, "http://domain.com", nil)) {
		t.Error("The comparisonTest must return false because we return the opposite of true (a < c)")
	}

	if !validateTest(logicalOrTest, httptest.NewRequest(http.MethodGet, "http://domain.com", nil)) {
		t.Error("The logicalOrTest must return true because we return true or false (1==1)|('abc'=='def')")
	}

	if validateTest(logicalAndTest, httptest.NewRequest(http.MethodGet, "http://domain.com", nil)) {
		t.Error("The logicalAndTest must return false because we return true and false (4!=5)&(4==5)")
	}

	rq := httptest.NewRequest(http.MethodGet, "http://domain.com", nil)
	rq.Header.Add("Accept-Language", "en-gb")
	rq.Header.Add("Accept-Language", "fr-fr")

	if !validateTest(complexTest, rq) {
		t.Error("The complexTest must return true")
	}
}
