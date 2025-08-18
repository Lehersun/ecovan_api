package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProblem_Serialization(t *testing.T) {
	problem := Problem{
		Type:     "/errors/test",
		Title:    "Test Error",
		Status:   400,
		Detail:   "This is a test error",
		Instance: "/test/endpoint",
	}

	data, err := json.Marshal(problem)
	if err != nil {
		t.Fatalf("Failed to marshal problem: %v", err)
	}

	var unmarshaled Problem
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal problem: %v", err)
	}

	if unmarshaled.Type != problem.Type {
		t.Errorf("Expected Type %s, got %s", problem.Type, unmarshaled.Type)
	}
	if unmarshaled.Title != problem.Title {
		t.Errorf("Expected Title %s, got %s", problem.Title, unmarshaled.Title)
	}
	if unmarshaled.Status != problem.Status {
		t.Errorf("Expected Status %d, got %d", problem.Status, unmarshaled.Status)
	}
	if unmarshaled.Detail != problem.Detail {
		t.Errorf("Expected Detail %s, got %s", problem.Detail, unmarshaled.Detail)
	}
	if unmarshaled.Instance != problem.Instance {
		t.Errorf("Expected Instance %s, got %s", problem.Instance, unmarshaled.Instance)
	}
}

func TestWriteProblem_Headers(t *testing.T) {
	problem := Problem{
		Type:   "/errors/test",
		Title:  "Test Error",
		Status: 400,
		Detail: "This is a test error",
	}

	recorder := httptest.NewRecorder()
	WriteProblem(recorder, problem)

	// Check Content-Type header
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/problem+json" {
		t.Errorf("Expected Content-Type 'application/problem+json', got '%s'", contentType)
	}

	// Check status code
	if recorder.Code != 400 {
		t.Errorf("Expected status code 400, got %d", recorder.Code)
	}
}

func TestWriteProblem_Serialization(t *testing.T) {
	problem := Problem{
		Type:   "/errors/test",
		Title:  "Test Error",
		Status: 400,
		Detail: "This is a test error",
	}

	recorder := httptest.NewRecorder()
	WriteProblem(recorder, problem)

	// Parse response body
	var response Problem
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response matches input
	if response.Type != problem.Type {
		t.Errorf("Expected Type %s, got %s", problem.Type, response.Type)
	}
	if response.Title != problem.Title {
		t.Errorf("Expected Title %s, got %s", problem.Title, response.Title)
	}
	if response.Status != problem.Status {
		t.Errorf("Expected Status %d, got %d", problem.Status, response.Status)
	}
	if response.Detail != problem.Detail {
		t.Errorf("Expected Detail %s, got %s", problem.Detail, response.Detail)
	}
}

func TestCommonProblems_StatusCodes(t *testing.T) {
	expectedStatuses := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusConflict,
		http.StatusUnsupportedMediaType,
		http.StatusUnprocessableEntity,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	}

	for _, status := range expectedStatuses {
		if problem, exists := CommonProblems[status]; !exists {
			t.Errorf("Expected problem for status %d", status)
		} else if problem.Status != status {
			t.Errorf("Expected problem status %d, got %d", status, problem.Status)
		}
	}
}

func TestWriteProblemWithDetail(t *testing.T) {
	recorder := httptest.NewRecorder()
	detail := "Custom error detail"

	WriteProblemWithDetail(recorder, http.StatusNotFound, detail)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, recorder.Code)
	}

	var response Problem
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Detail != detail {
		t.Errorf("Expected detail '%s', got '%s'", detail, response.Detail)
	}
}

func TestWriteProblemWithType(t *testing.T) {
	recorder := httptest.NewRecorder()
	customType := "/errors/custom"
	detail := "Custom error detail"

	WriteProblemWithType(recorder, http.StatusBadRequest, customType, detail)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var response Problem
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Type != customType {
		t.Errorf("Expected type '%s', got '%s'", customType, response.Type)
	}
	if response.Detail != detail {
		t.Errorf("Expected detail '%s', got '%s'", detail, response.Detail)
	}
}

func TestWriteCustomProblem(t *testing.T) {
	recorder := httptest.NewRecorder()
	customType := "/errors/custom"
	title := "Custom Title"
	status := 499
	detail := "Custom error detail"
	instance := "/custom/endpoint"

	WriteCustomProblem(recorder, customType, title, status, detail, instance)

	if recorder.Code != status {
		t.Errorf("Expected status code %d, got %d", status, recorder.Code)
	}

	var response Problem
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Type != customType {
		t.Errorf("Expected type '%s', got '%s'", customType, response.Type)
	}
	if response.Title != title {
		t.Errorf("Expected title '%s', got '%s'", title, response.Title)
	}
	if response.Status != status {
		t.Errorf("Expected status %d, got %d", status, response.Status)
	}
	if response.Detail != detail {
		t.Errorf("Expected detail '%s', got '%s'", detail, response.Detail)
	}
	if response.Instance != instance {
		t.Errorf("Expected instance '%s', got '%s'", instance, response.Instance)
	}
}

func TestHelperFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		helper   func(http.ResponseWriter, string)
		expected int
	}{
		{"WriteValidationError", WriteValidationError, http.StatusUnprocessableEntity},
		{"WriteNotFound", WriteNotFound, http.StatusNotFound},
		{"WriteConflict", WriteConflict, http.StatusConflict},
		{"WriteUnauthorized", WriteUnauthorized, http.StatusUnauthorized},
		{"WriteForbidden", WriteForbidden, http.StatusForbidden},
		{"WriteBadRequest", WriteBadRequest, http.StatusBadRequest},
		{"WriteInternalError", WriteInternalError, http.StatusInternalServerError},
		{"WriteServiceUnavailable", WriteServiceUnavailable, http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			detail := "Test detail"

			tc.helper(recorder, detail)

			if recorder.Code != tc.expected {
				t.Errorf("Expected status code %d, got %d", tc.expected, recorder.Code)
			}

			var response Problem
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Detail != detail {
				t.Errorf("Expected detail '%s', got '%s'", detail, response.Detail)
			}
		})
	}
}
