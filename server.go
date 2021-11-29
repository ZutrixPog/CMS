package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"

	"github.com/zutrixpog/CMS/graph"
	"github.com/zutrixpog/CMS/graph/generated"
	"github.com/zutrixpog/CMS/db"
	"github.com/zutrixpog/CMS/middleware"
)

var Database db.DB
const defaultPort = "8080"
const DB_URL = "mongodb+srv://zutrix:PgMing127001@erfancluster.r5bzv.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"

func init(){
	Database = db.DB{DbName: "myFirstDatabase"}
	Database.Connect(DB_URL)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	router := mux.NewRouter()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{DB: Database}}))
	router.Use(middleware.CookieMiddleware(&Database))
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
