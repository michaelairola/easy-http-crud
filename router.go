package easyHttpCrud

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

var corsOpts = []handlers.CORSOption{
	handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"}),
	handlers.AllowedOrigins([]string{"*"}),
	handlers.AllowedHeaders([]string{"authorization", "domain", "content-type"}),
	handlers.AllowCredentials(),
	handlers.OptionStatusCode(200),
}

// var METHODS = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
var METHODS = map[string]int{
	"GET":    1,
	"POST":   2,
	"PUT":    3,
	"PATCH":  4,
	"DELETE": 5,
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	AuthLevel   string
	HandlerFunc http.HandlerFunc
	Policy      mux.MiddlewareFunc
}
type Routes []Route

/*
	parseRouteMap converts easy-to-read string-handler maps
	into the Route structs defined above.
	ex:
	map[string]http.HandlerFunc{
		"AuthLevel Method Pattern ": handlerFunc
	}

	will be converted to

	Route{Name, Method, Pattern, AuthLevel, HandlerFunc, Policy}

*/
func parseRouteMap(line string, handler http.HandlerFunc) Route {
	token := strings.Split(line, " ")
	var name string
	var method string
	var pattern string
	var authLevel string
	var policy mux.MiddlewareFunc

	policyOrMethod := token[0]
	if METHODS[policyOrMethod] != 0 {
		authLevel = "SECURE"
	} else if POLICIES[policyOrMethod] != nil {
		authLevel = policyOrMethod
		token = token[1:]
	} else {
		fmt.Println("ROUTE_DEF_ERR: Route Definition must start with either an Authorized HTTP method, or a defined Policy in defie-service auth.go file")
		return Route{}
	}
	method = token[0]
	pattern = token[1]
	name = strings.Join(token[2:], " ")
	policy = POLICIES[authLevel]
	return Route{
		name, method, pattern, authLevel, handler, policy,
	}
}

// simple router that redirects no slash to slash
func initRouter() *mux.Router {
	return mux.NewRouter().StrictSlash(true)
}

// if you have multiple routeSpecs, this adds them together
func joinRouteMaps(routeSpecs ...map[string]http.HandlerFunc) Routes {
	routeMap := make(map[string]http.HandlerFunc)
	for _, route := range routeSpecs {
		for metaData, handler := range route {
			routeMap[metaData] = handler
		}
	}
	var routes []Route
	for k, v := range routeMap {
		route := parseRouteMap(k, v)
		routes = append(routes, route)
	}
	return routes
}

// addRoutes: add route structs to the given router :)
func addRoutes(router *mux.Router, service string, routes []Route) *mux.Router {
	for _, route := range routes {
		handler := route.Policy(route.HandlerFunc)
		router.
			Methods(route.Method).
			Path("/api/" + service + route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

func (r Routes) Len() int      { return len(r) }
func (r Routes) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r Routes) Less(i, j int) bool {
	method1 := METHODS[r[i].Method]
	method2 := METHODS[r[j].Method]
	if method1 == method2 {
		return r[i].Name < r[j].Name
	}
	return method1 < method2
}

func prettyPrintRoutes(routes Routes) {
	sort.Sort(routes)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.Debug|tabwriter.AlignRight)
	fmt.Fprintln(w, "AuthLevel\tMethod\t API\t Description")
	fmt.Fprintln(w, "---------\t------\t----\t------------")
	for _, route := range routes {
		fmt.Fprintln(w, route.AuthLevel, "\t", route.Method, "\t", route.Pattern, "\t", route.Name)
	}
	fmt.Println()
	w.Flush()
	fmt.Println()
}

// -------------------------------------------------------------------------
// Route Map Creators
// -------------------------------------------------------------------------

func createHealthRoute(Service string) map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"OPEN GET /health Check Response Time": func(w http.ResponseWriter, r *http.Request) {
			SendObject(w, 200, HealthResponse{HostName, Service, Version, "ok"})
		},
	}
}

/*	-----------------------------------------------------------------------
						crudRoutes!!!!!!!
	-----------------------------------------------------------------------
	This is a fully-functional, CRUD route-creator!
	current apis:
		* GET /search?{querystring}
			querystring parsed with rules found in utils.go in ParseQueryString
		* GET /single/{Id}
			this will get the model with Id from the table associated to that model
		* GET /history/{Id}
			assuming record has Id & Ver, this will get all history versions with associated Id
		* Get /history/{Id}/{Ver}
			assuming record has Id & Ver, this will get history record with specified Id and Ver
		* POST /create
			creates a new value
		* PUT /update/{Id}
			will update record with stated Id
		* DELETE /delete/{Id}
			deletes said record

	valid option values:
		* 'merge'defie.Merger
			this is the function used to merge to record values
		* "reports"
			gives ability to create report and save as a file.
*/
func createCrudRoutes(model interface{}, options map[string]interface{}) map[string]http.HandlerFunc {
	CRUDRoutes := make(map[string]http.HandlerFunc)
	p, _ := options["crud-permissions"].(string)
	Permission := CrudPermissions(p)
	// var Service string
	Service, ok := model.(string)
	if !ok {
		Service = TableNameOf(model)
		searchRoute := Permission + " GET /search Queries " + Service + " table"
		getByIdRoute := Permission + " GET /single/{Id} Get " + Service + " by Id"
		createRoute := Permission + " POST /create Create " + Service
		updateRoute := Permission + " PUT /update/{Id} Update " + Service
		deleteRoute := Permission + " DELETE /delete/{Id} Delete " + Service
		CRUDRoutes = map[string]http.HandlerFunc{
			searchRoute:  searchFactory(model, scopeOpts(options, "search")),
			getByIdRoute: getByIdFactory(model, scopeOpts(options, "single")),
			createRoute:  createFactory(model, scopeOpts(options, "create")),
			updateRoute:  updateFactory(model, scopeOpts(options, "update")),
			deleteRoute:  deleteFactory(model, scopeOpts(options, "delete")),
		}
		if hasHistory(model) {
			searchHistoryRoute := Permission + " GET /history/{Id} Queries " + Service + " history table"
			getHistByIdVer := Permission + " GET /history/{Id}/{Ver} Get" + Service + " history by Id and Version"
			CRUDRoutes[searchHistoryRoute] = searchHistFactory(model, scopeOpts(options, "history-search"))
			CRUDRoutes[getHistByIdVer] = getHistoryFactory(model, scopeOpts(options, "history"))
		}
	}
	// do optional things
	for i, val := range options {
		switch i {
		case "merge":
			key := TableNameOf(model)
			route := Permission + " PUT /merge/{Id1}/{Id2} Merge " + key + " 1 into " + key + " 2"
			CRUDRoutes[route] = mergeFactory(model, options)
		case "report":
			hasReports, ok := val.(bool)
			if ok && hasReports {
				createReportRoute := Permission + " POST /report create report for " + Service + " list page"
				CRUDRoutes[createReportRoute] = createReport(model, scopeOpts(options, "report"))
			}
		}

	}
	return CRUDRoutes
}

/*
	Any microservice initialized with this function will
	Have crud Routes, health routes, and the specialRoutes given!
*/

func DefieRouter(specialRoutes map[string]http.HandlerFunc, model interface{}, options map[string]interface{}) http.Handler {
	service := TableNameOf(model)
	fmt.Println()
	fmt.Println("Service:", service)
	fmt.Println("Version:", Version)
	fmt.Println()
	healthRoutes := createHealthRoute(service)
	crudRoutes := createCrudRoutes(model, options)
	routes := joinRouteMaps(healthRoutes, crudRoutes, specialRoutes)
	prettyPrintRoutes(routes)
	router := initRouter()
	router = addRoutes(router, service, routes)
	return handlers.CORS(corsOpts...)(router)
}
