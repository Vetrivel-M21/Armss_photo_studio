// Package dto holds request/response shapes for the HTTP API, kept separate from
// domain and storage types so wire format can change without touching either.
package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
