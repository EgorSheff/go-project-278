package main

import (
	"context"
	"log/slog"
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
		slog.Error("Unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	dao := generated.New(conn)

	r := gin.Default()

	handlers.RegisterHandlers(r, dao)

	if err := r.Run(":8080"); err != nil {
		slog.Error("Run HTTP server error", "err", err)
	}
}
