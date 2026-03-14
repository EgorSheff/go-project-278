package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/EgorSheff/go-project-278/db/generated"
	"github.com/EgorSheff/go-project-278/handlers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	dao := generated.New(conn)

	r := gin.Default()

	handlers.RegisterHandlers(r, dao)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
