package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	killgrave "github.com/friendsofgo/killgrave/internal"
	"github.com/friendsofgo/killgrave/internal/debugger"
)

func TestImposterHandler(t *testing.T) {
	bodyRequest := []byte(`{
		"data": {
			"type": "gophers",
		  "attributes": {
			"name": "Zebediah",
			"color": "Purple",
			"age": 55
		  }
		}
	  }`)
	var headers = make(map[string]string)
	headers["Content-Type"] = "application/json"

	schemaFile := "test/testdata/imposters/schemas/create_gopher_request.json"
	bodyFile := "test/testdata/imposters/responses/create_gopher_response.json"
	bodyFileFake := "test/testdata/imposters/responses/create_gopher_response_fail.json"
	body := `{"test":true}`

	validRequest := killgrave.Request{
		Method:     "POST",
		Endpoint:   "/gophers",
		SchemaFile: &schemaFile,
		Headers:    &headers,
	}

	f, _ := os.Open(bodyFile)
	defer f.Close()
	expectedBodyFileData, _ := ioutil.ReadAll(f)

	var dataTest = []struct {
		name         string
		imposter     killgrave.Imposter
		expectedBody string
		statusCode   int
	}{
		{"valid imposter with body", killgrave.Imposter{Request: validRequest, Response: killgrave.Response{Status: http.StatusOK, Headers: &headers, Body: body}}, body, http.StatusOK},
		{"valid imposter with bodyFile", killgrave.Imposter{Request: validRequest, Response: killgrave.Response{Status: http.StatusOK, Headers: &headers, BodyFile: &bodyFile}}, string(expectedBodyFileData), http.StatusOK},
		{"valid imposter with not exists bodyFile", killgrave.Imposter{Request: validRequest, Response: killgrave.Response{Status: http.StatusOK, Headers: &headers, BodyFile: &bodyFileFake}}, "", http.StatusOK},
	}

	for _, tt := range dataTest {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/gophers", bytes.NewBuffer(bodyRequest))
			if err != nil {
				t.Fatalf("could not created request: %v", err)
			}

			rec := httptest.NewRecorder()
			handler := imposterHandler(debugger.NewNoOp(), tt.imposter)

			handler.ServeHTTP(rec, req)
			if status := rec.Code; status != tt.statusCode {
				t.Errorf("handler expected %d code and got: %d code", tt.statusCode, status)
			}

			if rec.Body.String() != tt.expectedBody {
				t.Errorf("handler expected %s body and got: %s body", tt.expectedBody, rec.Body.String())
			}

		})
	}
}

func TestInvalidRequestWithSchema(t *testing.T) {
	validRequest := []byte(`{
		"data": {
			"type": "gophers",
		  "attributes": {
			"name": "Zebediah",
			"color": "Purple"
		  }
		}
	  }`)

	var dataTest = []struct {
		name       string
		imposter   killgrave.Imposter
		statusCode int
		request    []byte
	}{
		{"valid request no schema", killgrave.Imposter{Request: killgrave.Request{Method: "POST", Endpoint: "/gophers"}, Response: killgrave.Response{Status: http.StatusOK, Body: "test ok"}}, http.StatusOK, validRequest},
	}

	for _, tt := range dataTest {

		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/gophers", bytes.NewBuffer(tt.request))
			if err != nil {
				t.Fatalf("could not created request: %v", err)
			}
			rec := httptest.NewRecorder()
			handler := imposterHandler(debugger.NewNoOp(), tt.imposter)

			handler.ServeHTTP(rec, req)
			if status := rec.Code; status != tt.statusCode {
				t.Fatalf("handler expected %d code and got: %d code", tt.statusCode, status)
			}
		})
	}
}
