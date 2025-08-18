package http

import (
	"encoding/json"
	"net/http"
)

// Problem represents an HTTP Problem Details object as defined in RFC 7807
type Problem struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// Common problem types
const (
	ProblemTypeInvalidInput        = "/errors/invalid-input"
	ProblemTypeUnauthorized        = "/errors/unauthorized"
	ProblemTypeForbidden           = "/errors/forbidden"
	ProblemTypeNotFound            = "/errors/not-found"
	ProblemTypeConflict            = "/errors/conflict"
	ProblemTypeValidationError     = "/errors/validation-error"
	ProblemTypeInternalError       = "/errors/internal-error"
	ProblemTypeServiceUnavailable  = "/errors/service-unavailable"
	ProblemTypeBadRequest          = "/errors/bad-request"
	ProblemTypeMethodNotAllowed    = "/errors/method-not-allowed"
	ProblemTypeUnsupportedMediaType = "/errors/unsupported-media-type"
)

// Common problems for standard HTTP status codes
var CommonProblems = map[int]Problem{
	http.StatusBadRequest: {
		Type:   ProblemTypeBadRequest,
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
	},
	http.StatusUnauthorized: {
		Type:   ProblemTypeUnauthorized,
		Title:  "Unauthorized",
		Status: http.StatusUnauthorized,
	},
	http.StatusForbidden: {
		Type:   ProblemTypeForbidden,
		Title:  "Forbidden",
		Status: http.StatusForbidden,
	},
	http.StatusNotFound: {
		Type:   ProblemTypeNotFound,
		Title:  "Not Found",
		Status: http.StatusNotFound,
	},
	http.StatusMethodNotAllowed: {
		Type:   ProblemTypeMethodNotAllowed,
		Title:  "Method Not Allowed",
		Status: http.StatusMethodNotAllowed,
	},
	http.StatusConflict: {
		Type:   ProblemTypeConflict,
		Title:  "Conflict",
		Status: http.StatusConflict,
	},
	http.StatusUnsupportedMediaType: {
		Type:   ProblemTypeUnsupportedMediaType,
		Title:  "Unsupported Media Type",
		Status: http.StatusUnsupportedMediaType,
	},
	http.StatusUnprocessableEntity: {
		Type:   ProblemTypeValidationError,
		Title:  "Validation Error",
		Status: http.StatusUnprocessableEntity,
	},
	http.StatusInternalServerError: {
		Type:   ProblemTypeInternalError,
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
	},
	http.StatusServiceUnavailable: {
		Type:   ProblemTypeServiceUnavailable,
		Title:  "Service Unavailable",
		Status: http.StatusServiceUnavailable,
	},
}

// WriteProblem writes a Problem JSON response to the HTTP response writer
func WriteProblem(w http.ResponseWriter, p Problem) {
	// Set content type header
	w.Header().Set("Content-Type", "application/problem+json")
	
	// Set status code
	w.WriteHeader(p.Status)
	
	// Encode and write the problem
	if err := json.NewEncoder(w).Encode(p); err != nil {
		// Fallback to a simple error if JSON encoding fails
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// WriteProblemWithDetail creates and writes a problem with custom detail
func WriteProblemWithDetail(w http.ResponseWriter, status int, detail string) {
	problem := CommonProblems[status]
	problem.Detail = detail
	WriteProblem(w, problem)
}

// WriteProblemWithType creates and writes a problem with custom type and detail
func WriteProblemWithType(w http.ResponseWriter, status int, problemType, detail string) {
	problem := CommonProblems[status]
	problem.Type = problemType
	problem.Detail = detail
	WriteProblem(w, problem)
}

// WriteCustomProblem creates and writes a completely custom problem
func WriteCustomProblem(w http.ResponseWriter, problemType, title string, status int, detail, instance string) {
	problem := Problem{
		Type:     problemType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
	WriteProblem(w, problem)
}

// WriteValidationError writes a validation error problem
func WriteValidationError(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusUnprocessableEntity, detail)
}

// WriteNotFound writes a not found problem
func WriteNotFound(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusNotFound, detail)
}

// WriteConflict writes a conflict problem
func WriteConflict(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusConflict, detail)
}

// WriteUnauthorized writes an unauthorized problem
func WriteUnauthorized(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusUnauthorized, detail)
}

// WriteForbidden writes a forbidden problem
func WriteForbidden(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusForbidden, detail)
}

// WriteBadRequest writes a bad request problem
func WriteBadRequest(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusBadRequest, detail)
}

// WriteInternalError writes an internal server error problem
func WriteInternalError(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusInternalServerError, detail)
}

// WriteServiceUnavailable writes a service unavailable problem
func WriteServiceUnavailable(w http.ResponseWriter, detail string) {
	WriteProblemWithDetail(w, http.StatusServiceUnavailable, detail)
}
