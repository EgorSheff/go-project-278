package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/EgorSheff/go-project-278/db/generated"
	"github.com/EgorSheff/go-project-278/dto"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUnsupportedRangeFormat = errors.New("unsupported range format")
)

type apiHandlers struct {
	dao     *generated.Queries
	baseURL *url.URL
}

func RegisterHandlers(g *gin.RouterGroup, dao *generated.Queries) {
	baseURL, err := url.Parse(os.Getenv("BASE_URL"))
	if err != nil {
		return
	}

	apiH := apiHandlers{dao: dao, baseURL: baseURL}

	g.GET("links", apiH.getLinksList)
	g.POST("links", apiH.createLink)
	g.GET("links/:id", apiH.getLink)
	g.PUT("links/:id", apiH.updateLink)
	g.DELETE("links/:id", apiH.deleteLink)

	g.GET("link_visits", apiH.getLinkVisits)
}

func (a *apiHandlers) getLinksList(c *gin.Context) {
	offset, limit, err := parseRange(c)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}

	links, err := a.dao.ListLinks(context.Background(), generated.ListLinksParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	count, err := a.dao.CountLinks(context.Background())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", offset, limit, count))

	result := make([]dto.Link, 0, len(links))
	for _, link := range links {
		result = append(result, dto.Link{
			Id:          link.ID,
			OriginalURL: link.OriginalUrl,
			ShortName:   link.ShortName,
			ShortURL:    a.baseURL.ResolveReference(&url.URL{Path: "/r/" + link.ShortName}).String(),
		})
	}
	c.JSON(http.StatusOK, result)
}

func (a *apiHandlers) createLink(c *gin.Context) {
	var req dto.Link
	if err := c.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": validationErrors(ve)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	shortName := req.ShortName
	if shortName == "" {
		shortName = strings.ReplaceAll(uuid.Must(uuid.NewV4()).String(), "-", "")
	}

	link, err := a.dao.CreateLink(context.Background(), generated.CreateLinkParams{
		OriginalUrl: req.OriginalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": "short name already in use"}})
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	resp := dto.Link{
		Id:          link.ID,
		OriginalURL: link.OriginalUrl,
		ShortName:   link.ShortName,
		ShortURL:    a.baseURL.ResolveReference(&url.URL{Path: "/r/" + link.ShortName}).String(),
	}

	c.JSON(http.StatusCreated, resp)
}

func (a *apiHandlers) getLink(c *gin.Context) {
	idNum, err := getId(c)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}

	link, err := a.dao.GetLink(context.Background(), int32(idNum))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, link)
}

func (a *apiHandlers) updateLink(c *gin.Context) {
	idNum, err := getId(c)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}

	var req dto.Link
	if err := c.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": validationErrors(ve)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	shortName := req.ShortName
	if shortName == "" {
		shortName = strings.ReplaceAll(uuid.Must(uuid.NewV4()).String(), "-", "")
	}

	if _, err := a.dao.UpdateLink(context.Background(), generated.UpdateLinkParams{
		ID:          int32(idNum),
		OriginalUrl: req.OriginalURL,
		ShortName:   shortName,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
			return
		}
		if isUniqueViolation(err) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": "short name already in use"}})
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func (a *apiHandlers) deleteLink(c *gin.Context) {
	idNum, err := getId(c)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}

	if err := a.dao.DeleteLink(context.Background(), int32(idNum)); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *apiHandlers) getLinkVisits(c *gin.Context) {
	offset, limit, err := parseRange(c)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}

	visits, err := a.dao.ListVisits(context.Background(), generated.ListVisitsParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	count, err := a.dao.CountVisits(context.Background())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", offset, limit, count))

	result := make([]dto.Visit, 0, len(visits))
	for _, visit := range visits {
		result = append(result, dto.Visit{
			ID:        visit.ID,
			LinkID:    visit.LinkID,
			CreatedAt: visit.CreatedAt.Time,
			Ip:        visit.Ip,
			UserAgent: visit.UserAgent,
			Status:    visit.Status,
		})
	}
	c.JSON(http.StatusOK, result)
}

func getId(c *gin.Context) (int32, error) {
	id := c.Param("id")
	idNum, err := strconv.Atoi(id)
	return int32(idNum), err
}

var fieldNames = map[string]string{
	"OriginalURL": "original_url",
	"ShortName":   "short_name",
}

func validationErrors(ve validator.ValidationErrors) gin.H {
	errs := gin.H{}
	for _, fe := range ve {
		field := fieldNames[fe.Field()]
		errs[field] = fe.Error()
	}
	return errs
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var rangeRegexp = regexp.MustCompile(`\[(\d+),(\d+)\]`)

func parseRange(c *gin.Context) (offset, limit int32, err error) {
	raw := c.Query("range")
	if raw == "" {
		if c.GetHeader("Range") != "" {
			raw = strings.ReplaceAll(c.GetHeader("Range"), " ", "")
		} else {
			return 0, 100, nil
		}
	}

	match := rangeRegexp.FindStringSubmatch(raw)
	if len(match) != 3 {
		fmt.Println(match, raw, c.Request.URL.Query())
		return 0, 0, ErrUnsupportedRangeFormat
	}
	off, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, 0, err
	}
	lim, err := strconv.Atoi(match[2])
	if err != nil {
		return 0, 0, err
	}

	if lim > 100 || lim == 0 {
		lim = 100
	}

	return int32(off), int32(lim), nil
}
