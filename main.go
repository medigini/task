package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"task/internal/api/graphql/graph"
	"task/internal/services"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/allegro/bigcache/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {

	app := fiber.New()
	cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(600*time.Minute))
	svc := services.GraphQLService{
		Cache: cache,
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		Svc: &svc,
	}}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	// graphql handler

	gqlhandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.ServeHTTP(w, r)

	})
	playground := playground.Handler("GraphQL playground", "/graphql")
	// mounting graphql handler in fiber
	app.All("/graphql", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(gqlhandler)(c.Context())
		return nil
	})

	app.Get("/", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(playground)(c.Context())
		return nil
	})

	fmt.Println("🚀 Server running at http://localhost:4000/ ")
	log.Fatal(app.Listen(":4000"))

}
