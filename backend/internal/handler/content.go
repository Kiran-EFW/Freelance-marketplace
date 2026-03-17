package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ContentArticle mirrors the service-level Article type for handler use.
type ContentArticle struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Body        string    `json:"body"`
	Category    string    `json:"category"`
	Audience    string    `json:"audience"`
	Tags        []string  `json:"tags"`
	Language    string    `json:"language"`
	AuthorName  string    `json:"author_name"`
	IsPublished bool      `json:"is_published"`
	ViewCount   int       `json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContentServiceInterface defines the business operations required by ContentHandler.
type ContentServiceInterface interface {
	ListArticles(ctx context.Context, audience, category, lang string, limit, offset int) ([]ContentArticle, int, error)
	GetArticle(ctx context.Context, slug string) (*ContentArticle, error)
	GetArticleByID(ctx context.Context, id uuid.UUID) (*ContentArticle, error)
	CreateArticle(ctx context.Context, article *ContentArticle) error
	UpdateArticle(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteArticle(ctx context.Context, id uuid.UUID) error
	GetPopular(ctx context.Context, audience string, limit int) ([]ContentArticle, error)
	GetRelated(ctx context.Context, articleID uuid.UUID, limit int) ([]ContentArticle, error)
}

// ContentHandler handles content and education endpoints.
type ContentHandler struct {
	service ContentServiceInterface
}

// NewContentHandler creates a new ContentHandler.
func NewContentHandler(svc ContentServiceInterface) *ContentHandler {
	return &ContentHandler{service: svc}
}

// RegisterRoutes mounts public content routes on the given Fiber router group.
// These endpoints do not require authentication.
func (h *ContentHandler) RegisterRoutes(rg fiber.Router) {
	rg.Get("/", h.ListArticles)
	rg.Get("/popular", h.GetPopularArticles)
	rg.Get("/:slug", h.GetArticleBySlug)
	rg.Get("/:id/related", h.GetRelatedArticles)
}

// RegisterAdminRoutes mounts admin content routes on the given Fiber router group.
// These endpoints require admin authentication.
func (h *ContentHandler) RegisterAdminRoutes(rg fiber.Router) {
	rg.Post("/", h.CreateArticle)
	rg.Put("/:id", h.UpdateArticle)
	rg.Delete("/:id", h.DeleteArticle)
}

// ListArticles returns a paginated list of published articles.
// GET /api/v1/content?audience=customer&category=customer_tip&language=en&page=1&per_page=20
func (h *ContentHandler) ListArticles(c *fiber.Ctx) error {
	audience := c.Query("audience", "")
	category := c.Query("category", "")
	language := c.Query("language", "")

	page := 1
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	perPage := 20
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			perPage = v
		}
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	articles, total, err := h.service.ListArticles(c.UserContext(), audience, category, language, perPage, offset)
	if err != nil {
		log.Error().Err(err).Msg("content: failed to list articles")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to list articles",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": articles,
		"meta": fiber.Map{
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	})
}

// GetArticleBySlug returns a single published article by slug.
// GET /api/v1/content/:slug
func (h *ContentHandler) GetArticleBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "slug is required",
			},
		})
	}

	article, err := h.service.GetArticle(c.UserContext(), slug)
	if err != nil {
		log.Error().Err(err).Str("slug", slug).Msg("content: failed to get article")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve article",
			},
		})
	}

	if article == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "article not found",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": article,
	})
}

// GetPopularArticles returns the most-viewed articles.
// GET /api/v1/content/popular?audience=customer&limit=10
func (h *ContentHandler) GetPopularArticles(c *fiber.Ctx) error {
	audience := c.Query("audience", "")

	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > 50 {
		limit = 50
	}

	articles, err := h.service.GetPopular(c.UserContext(), audience, limit)
	if err != nil {
		log.Error().Err(err).Msg("content: failed to get popular articles")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve popular articles",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": articles,
	})
}

// GetRelatedArticles returns articles related to the given article ID.
// GET /api/v1/content/:id/related?limit=5
func (h *ContentHandler) GetRelatedArticles(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid article ID format",
			},
		})
	}

	limit := 5
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	articles, err := h.service.GetRelated(c.UserContext(), id, limit)
	if err != nil {
		log.Error().Err(err).Str("article_id", id.String()).Msg("content: failed to get related articles")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve related articles",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": articles,
	})
}

// createArticleRequest is the payload for POST /api/v1/admin/content.
type createArticleRequest struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Body        string   `json:"body"`
	Category    string   `json:"category"`
	Audience    string   `json:"audience"`
	Tags        []string `json:"tags"`
	Language    string   `json:"language"`
	AuthorName  string   `json:"author_name"`
	IsPublished bool     `json:"is_published"`
}

func (r *createArticleRequest) validate() error {
	if r.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	if r.Category == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category is required")
	}
	if r.Audience == "" {
		return fiber.NewError(fiber.StatusBadRequest, "audience is required")
	}
	return nil
}

// CreateArticle creates a new article (admin only).
// POST /api/v1/admin/content
func (h *ContentHandler) CreateArticle(c *fiber.Ctx) error {
	var req createArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := req.validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
	}

	article := &ContentArticle{
		Title:       req.Title,
		Summary:     req.Summary,
		Body:        req.Body,
		Category:    req.Category,
		Audience:    req.Audience,
		Tags:        req.Tags,
		Language:    req.Language,
		AuthorName:  req.AuthorName,
		IsPublished: req.IsPublished,
	}

	if err := h.service.CreateArticle(c.UserContext(), article); err != nil {
		log.Error().Err(err).Str("title", req.Title).Msg("content: failed to create article")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to create article",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": article,
	})
}

// updateArticleRequest is the payload for PUT /api/v1/admin/content/:id.
type updateArticleRequest struct {
	Title       *string  `json:"title,omitempty"`
	Slug        *string  `json:"slug,omitempty"`
	Summary     *string  `json:"summary,omitempty"`
	Body        *string  `json:"body,omitempty"`
	Category    *string  `json:"category,omitempty"`
	Audience    *string  `json:"audience,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Language    *string  `json:"language,omitempty"`
	AuthorName  *string  `json:"author_name,omitempty"`
	IsPublished *bool    `json:"is_published,omitempty"`
}

// UpdateArticle updates an existing article (admin only).
// PUT /api/v1/admin/content/:id
func (h *ContentHandler) UpdateArticle(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid article ID format",
			},
		})
	}

	var req updateArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Audience != nil {
		updates["audience"] = *req.Audience
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}
	if req.AuthorName != nil {
		updates["author_name"] = *req.AuthorName
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "no fields to update",
			},
		})
	}

	if err := h.service.UpdateArticle(c.UserContext(), id, updates); err != nil {
		log.Error().Err(err).Str("article_id", id.String()).Msg("content: failed to update article")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to update article",
			},
		})
	}

	// Fetch and return the updated article.
	updated, err := h.service.GetArticleByID(c.UserContext(), id)
	if err != nil || updated == nil {
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"id":      id,
				"message": "article updated successfully",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": updated,
	})
}

// DeleteArticle deletes an article (admin only).
// DELETE /api/v1/admin/content/:id
func (h *ContentHandler) DeleteArticle(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_ID",
				"message": "invalid article ID format",
			},
		})
	}

	if err := h.service.DeleteArticle(c.UserContext(), id); err != nil {
		log.Error().Err(err).Str("article_id", id.String()).Msg("content: failed to delete article")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "failed to delete article",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":      id,
			"message": "article deleted successfully",
		},
	})
}
