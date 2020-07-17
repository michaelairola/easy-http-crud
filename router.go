package easyHttpCrud

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	PATCH  Method = "PATCH"
	DELETE Method = "DELETE"
)

type Route struct {
	Name     string
	Method   Method
	Pathname string
	Handler  http.HandlerFunc
	// Policy      mux.MiddlewareFunc
	Description string
}
type Routes []Route

type RouterOpts struct {
	DbConn      string
	ExtraRoutes Routes
}

type SearchRoute struct{}
type GetByIdRoute struct{}
type CreateRoute struct{}
type UpdateRoute struct{}
type DeleteRoute struct{}

func (search SearchRoute) Constructor(model interface{}, options map[string]interface{}) Route {
	Service := TableNameOf(model)
	return Route{
		Name:        Service + "-search",
		Method:      GET,
		Pathname:    "/search",
		Handler:     searchFactory(model, options),
		Description: "Queries " + Service + " table.",
	}
}

func (getById GetByIdRoute) Constructor(model interface{}, options map[string]interface{}) Route {
	Service := TableNameOf(model)
	return Route{
		Name:        Service + "-by_id",
		Method:      GET,
		Pathname:    "/{Id}",
		Handler:     getByIdFactory(model, options),
		Description: "Get " + Service + " by Id",
	}
}
func (create CreateRoute) Constructor(model interface{}, options map[string]interface{}) Route {
	Service := TableNameOf(model)
	return Route{
		Name:        Service + "-create",
		Method:      POST,
		Pathname:    "/create",
		Handler:     createFactory(model, options),
		Description: "Create a" + Service + " record.",
	}
}
func (create UpdateRoute) Constructor(model interface{}, options map[string]interface{}) Route {
	Service := TableNameOf(model)
	return Route{
		Name:        Service + "-update",
		Method:      PUT,
		Pathname:    "/update/{Id}",
		Handler:     updateFactory(model, options),
		Description: "Update " + Service + " record with id.",
	}
}
func (create DeleteRoute) Constructor(model interface{}, options map[string]interface{}) Route {
	Service := TableNameOf(model)
	return Route{
		Name:        Service + "",
		Method:      PUT,
		Pathname:    "/delete/{Id}",
		Handler:     deleteFactory(model, options),
		Description: "Delete " + Service + " record with id.",
	}
}

type RouteConstructor interface {
	Constructor(interface{}, map[string]interface{}) Route
}

var crudRoutes = []RouteConstructor{
	HealthRoute{},
	SearchRoute{},
	GetByIdRoute{},
	CreateRoute{},
	UpdateRoute{},
	DeleteRoute{},
}

/*
	Any microservice initialized with this function will
	Have crud Routes, health routes, and the specialRoutes given!
*/

func CreateRouter(model interface{}, options map[string]interface{}) http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	service := TableNameOf(model)
	for _, crudRoute := range crudRoutes {
		route := crudRoute.Constructor(model, options)
		handler := route.Handler
		router.
			Methods(string(route.Method)).
			Path("/api/" + service + route.Pathname).
			Name(route.Name).
			Handler(handler)
	}
	return router
	// return handlers(router)
}

// var corsOpts = []handlers.CORSOption{
// 	handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"}),
// 	handlers.AllowedOrigins([]string{"*"}),
// 	handlers.AllowedHeaders([]string{"authorization", "domain", "content-type"}),
// 	handlers.AllowCredentials(),
// 	handlers.OptionStatusCode(200),
// }
// return handlers /*.CORS(corsOpts...)*/ (router)
