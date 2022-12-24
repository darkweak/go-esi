package esi_test

import (
	"fmt"
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

func Test_Parse_includeMock(t *testing.T) {
	t.Parallel()
	fmt.Println(string(esi.Parse(loadFromFixtures("include"), httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_commentMock(t *testing.T) {
	t.Parallel()
	fmt.Println(string(esi.Parse(loadFromFixtures("comment"), httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_chooseMock(t *testing.T) {
	t.Parallel()
	fmt.Println(string(esi.Parse(loadFromFixtures("choose"), httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_escapeMock(t *testing.T) {
	t.Parallel()
	fmt.Println(string(esi.Parse(loadFromFixtures("escape"), httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_removeMock(t *testing.T) {
	t.Parallel()
	fmt.Println(string(esi.Parse(loadFromFixtures("remove"), httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_varsMock(t *testing.T) {
	t.Parallel()
	fmt.Println(string(esi.Parse(loadFromFixtures("vars"), httptest.NewRequest(http.MethodGet, "/", nil))))
}

func Test_Parse_fullMock(t *testing.T) {
	t.Parallel()
	fmt.Println(string(esi.Parse(loadFromFixtures("full.html"), httptest.NewRequest(http.MethodGet, "/", nil))))
}
