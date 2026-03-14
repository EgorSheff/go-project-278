package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/EgorSheff/go-project-278/db/generated"
	"github.com/EgorSheff/go-project-278/handlers/api"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func RegisterHandlers(r *gin.Engine, dao *generated.Queries) {
	apiGroup := r.Group("/api/")
	api.RegisterHandlers(apiGroup, dao)

	baseURL, err := url.Parse(os.Getenv("BASE_URL"))
	if err != nil {
		return
	}
	r.GET("/r/:code", redirectHandler(dao, baseURL))
}

func redirectHandler(dao *generated.Queries, baseURL *url.URL) func(c *gin.Context) {
	return func(c *gin.Context) {
		code := c.Param("code")

		link, err := dao.GetLinkByShortName(context.Background(), code)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.Status(http.StatusNotFound)
				return
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if _, err := dao.CreateVisit(context.Background(), generated.CreateVisitParams{
			LinkID:    link.ID,
			Ip:        c.ClientIP(),
			UserAgent: c.GetHeader("user-agent"),
			Status:    http.StatusTemporaryRedirect,
		}); err != nil {
			slog.Error("CreateVisit DB error", "err", err)
		}

		c.Header("Referer", baseURL.String())
		c.Redirect(http.StatusTemporaryRedirect, link.OriginalUrl)
	}
}
