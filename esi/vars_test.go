package esi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_parseVars(t *testing.T) {
	parseVariables(logicalAndTest, httptest.NewRequest(http.MethodGet, "http://domain.com", nil))
}
