package esi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_parseVars(t *testing.T) {
	t.Parallel()
	parseVariables(logicalAndTest, httptest.NewRequest(http.MethodGet, "http://domain.com", nil))
}
