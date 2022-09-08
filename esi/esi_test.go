package esi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func loadFromFixtures(name string) []byte {
	b, e := os.ReadFile("../fixtures/" + name)
	if e != nil {
		panic("The file " + name + " doesn't exist.")
	}

	return b
}

var (
	commentMock = loadFromFixtures("comment")
	chooseMock  = loadFromFixtures("choose")
	escapeMock  = loadFromFixtures("escape")
	includeMock = loadFromFixtures("include")
	removeMock  = loadFromFixtures("remove")
	// tryMock     = loadFromFixtures("try")
	varsMock = loadFromFixtures("vars")
)

func Test_Parse_includeMock(t *testing.T) {
	fmt.Println(string(Parse(includeMock, httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_commentMock(t *testing.T) {
	fmt.Println(string(Parse(commentMock, httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_chooseMock(t *testing.T) {
	fmt.Println(string(Parse(chooseMock, httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_escapeMock(t *testing.T) {
	fmt.Println(string(Parse(escapeMock, httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_removeMock(t *testing.T) {
	fmt.Println(string(Parse(removeMock, httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_varsMock(t *testing.T) {
	fmt.Println(string(Parse(varsMock, httptest.NewRequest(http.MethodGet, "/", nil))))
}
