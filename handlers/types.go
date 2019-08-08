package handlers

import "github.com/keratin/authn-server/app/services"

// This file contains some view model types (i.e. arguments and results consumed by handlers - swagger and JSON annotated)
// that are shared between handlers. Types specific to the handler are best placed in the same file as the handler.

type InputArgs interface {
	Validate() error
}

// ServiceErrors mirrors handlers.ServiceErrors
// swagger:response serviceErrors
type ServiceErrors struct {
	// in: body
	Errors services.FieldErrors `json:"errors"`
}

