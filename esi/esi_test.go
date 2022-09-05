package esi

import (
	"fmt"
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
	// chooseMock  = loadFromFixtures("choose")
	// escapeMock  = loadFromFixtures("escape")
	includeMock = loadFromFixtures("include")
	removeMock  = loadFromFixtures("remove")
	// tryMock     = loadFromFixtures("try")
	// varsMock    = loadFromFixtures("vars")
)

func Test_Parse_includeMock(t *testing.T) {
	fmt.Println(string(commentMock))
	fmt.Println(string(includeMock))
	fmt.Println(string(removeMock))
	// x := Parse(removeMock, httptest.NewRequest(http.MethodGet, "/", nil))
	// t.Error(string(x))
}
