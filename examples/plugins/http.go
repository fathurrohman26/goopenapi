package main

import (
	"log"
	"net/http"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
	"github.com/fathurrohman26/yaswag/pkg/yahttp"
)

func main() {
	// Create your OpenAPI spec
	spec := &openapi.Document{
		OpenAPI: "3.0.3",
		Info: openapi.Info{
			Title:       "Example API",
			Version:     "1.0.0",
			Description: "A sample API to demonstrate the HTTP middleware plugin",
		},
		Servers: []openapi.Server{
			{URL: "http://localhost:8080", Description: "Local server"},
		},
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Get: &openapi.Operation{
					Summary:     "List all users",
					OperationID: "listUsers",
					Tags:        []string{"users"},
					Parameters: []*openapi.Parameter{
						{
							Name:     "limit",
							In:       openapi.ParameterInQuery,
							Required: false,
							Schema:   openapi.IntegerSchema(),
						},
					},
					Responses: map[string]*openapi.Response{
						"200": {Description: "List of users"},
					},
				},
			},
			"/users/{id}": {
				Get: &openapi.Operation{
					Summary:     "Get user by ID",
					OperationID: "getUser",
					Tags:        []string{"users"},
					Parameters: []*openapi.Parameter{
						{
							Name:     "id",
							In:       openapi.ParameterInPath,
							Required: true,
							Schema:   openapi.IntegerSchema(),
						},
					},
					Responses: map[string]*openapi.Response{
						"200": {Description: "User found"},
						"404": {Description: "User not found"},
					},
				},
			},
		},
	}

	// Create HTTP middleware plugin using fluent builder
	handler := yahttp.WithSpec(spec).
		SpecPath("/openapi.json").
		SwaggerUIPath("/docs").
		EnableCORS().
		EnableLogging().
		EnableValidation().
		Mount(createRoutes())

	log.Println("Server starting on http://localhost:8080")
	log.Println("Swagger UI available at http://localhost:8080/docs")
	log.Println("OpenAPI spec available at http://localhost:8080/openapi.json")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

func createRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", usersHandler)
	mux.HandleFunc("/users/", userByIDHandler)
	return mux
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`[{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]`))
}

func userByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"id": 1, "name": "John"}`))
}
