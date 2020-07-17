package easyHttpCrud

import (
	"net/http"
)

type HealthResponse struct {
	HostName string
	Service  string
	Version  string
}

type HealthRoute struct{}

func (health HealthRoute) Constructor(model interface{}, options map[string]interface{}) Route {
	Service := TableNameOf(model)
	HealthHandler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SendObject(w, 200, HealthResponse{HostName, Service, Version})
	}))
	return Route{
		Name:        Service + "-health",
		Method:      GET,
		Pathname:    "/health",
		Handler:     HealthHandler,
		Description: "health endpoint for checking if service is functional",
	}
}
