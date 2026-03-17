package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog/log"
)

const (
	indexProviders  = "providers"
	indexCategories = "categories"
)

// SearchProvider defines the interface for search operations.
type SearchProvider interface {
	IndexProvider(ctx context.Context, provider ProviderDocument) error
	IndexCategory(ctx context.Context, category CategoryDocument) error
	SearchProviders(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error)
	SearchCategories(ctx context.Context, query string) ([]CategoryDocument, error)
	DeleteProvider(ctx context.Context, providerID string) error
	Setup() error
}

// ProviderDocument represents a searchable provider entry in Meilisearch.
type ProviderDocument struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Skills     []string `json:"skills"`
	Postcode   string   `json:"postcode"`
	Location   string   `json:"location"`
	Rating     float64  `json:"rating"`
	Category   string   `json:"category"`
	CategoryID string   `json:"category_id"`
	Language   []string `json:"language"`
	Geo        *GeoPoint `json:"_geo,omitempty"`
}

// GeoPoint represents a geographic coordinate for geo-based search.
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CategoryDocument represents a searchable category entry in Meilisearch.
type CategoryDocument struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Slug        string   `json:"slug"`
	Keywords    []string `json:"keywords"`
	ParentID    string   `json:"parent_id,omitempty"`
}

// SearchFilters holds the filter criteria for provider search.
type SearchFilters struct {
	Postcode   string  `json:"postcode,omitempty"`
	CategoryID string  `json:"category_id,omitempty"`
	Lat        float64 `json:"lat,omitempty"`
	Lng        float64 `json:"lng,omitempty"`
	RadiusKM   float64 `json:"radius_km,omitempty"`
	MinRating  float64 `json:"min_rating,omitempty"`
}

// SearchResults holds the results of a provider search.
type SearchResults struct {
	Hits             []ProviderDocument `json:"hits"`
	Total            int64              `json:"total"`
	ProcessingTimeMs int64              `json:"processing_time_ms"`
}

// SearchClient wraps the Meilisearch SDK client for provider and category search.
type SearchClient struct {
	client meilisearch.ServiceManager
}

// NewSearchClient creates a new Meilisearch client.
func NewSearchClient(host, apiKey string) *SearchClient {
	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))

	return &SearchClient{
		client: client,
	}
}

// Setup creates the Meilisearch indexes with proper settings for filterable,
// sortable, and searchable attributes.
func (s *SearchClient) Setup() error {
	// Create the providers index.
	_, err := s.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        indexProviders,
		PrimaryKey: "id",
	})
	if err != nil {
		log.Warn().Err(err).Msg("providers index may already exist")
	}

	providerIndex := s.client.Index(indexProviders)

	// Configure searchable attributes.
	_, err = providerIndex.UpdateSearchableAttributes(&[]string{
		"name", "skills", "location", "category",
	})
	if err != nil {
		return fmt.Errorf("search setup searchable attributes: %w", err)
	}

	// Configure filterable attributes.
	filterableAttrs := []interface{}{"postcode", "category_id", "rating", "language", "_geo"}
	_, err = providerIndex.UpdateFilterableAttributes(&filterableAttrs)
	if err != nil {
		return fmt.Errorf("search setup filterable attributes: %w", err)
	}

	// Configure sortable attributes.
	_, err = providerIndex.UpdateSortableAttributes(&[]string{
		"rating", "_geo",
	})
	if err != nil {
		return fmt.Errorf("search setup sortable attributes: %w", err)
	}

	// Create the categories index.
	_, err = s.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        indexCategories,
		PrimaryKey: "id",
	})
	if err != nil {
		log.Warn().Err(err).Msg("categories index may already exist")
	}

	categoryIndex := s.client.Index(indexCategories)

	_, err = categoryIndex.UpdateSearchableAttributes(&[]string{
		"name", "description", "keywords", "slug",
	})
	if err != nil {
		return fmt.Errorf("search setup category searchable attributes: %w", err)
	}

	log.Info().Msg("meilisearch indexes configured successfully")

	return nil
}

// IndexProvider adds or updates a provider document in the search index.
func (s *SearchClient) IndexProvider(_ context.Context, provider ProviderDocument) error {
	index := s.client.Index(indexProviders)

	_, err := index.AddDocuments([]ProviderDocument{provider}, nil)
	if err != nil {
		return fmt.Errorf("search index provider %s: %w", provider.ID, err)
	}

	log.Debug().Str("provider_id", provider.ID).Msg("provider indexed in meilisearch")

	return nil
}

// IndexCategory adds or updates a category document in the search index.
func (s *SearchClient) IndexCategory(_ context.Context, category CategoryDocument) error {
	index := s.client.Index(indexCategories)

	_, err := index.AddDocuments([]CategoryDocument{category}, nil)
	if err != nil {
		return fmt.Errorf("search index category %s: %w", category.ID, err)
	}

	log.Debug().Str("category_id", category.ID).Msg("category indexed in meilisearch")

	return nil
}

// SearchProviders searches for providers matching the query and filters.
func (s *SearchClient) SearchProviders(_ context.Context, query string, filters SearchFilters) (*SearchResults, error) {
	index := s.client.Index(indexProviders)

	searchReq := &meilisearch.SearchRequest{
		Limit: 20,
	}

	// Build filter expressions.
	var filterParts []string

	if filters.Postcode != "" {
		filterParts = append(filterParts, fmt.Sprintf("postcode = %q", filters.Postcode))
	}

	if filters.CategoryID != "" {
		filterParts = append(filterParts, fmt.Sprintf("category_id = %q", filters.CategoryID))
	}

	if filters.MinRating > 0 {
		filterParts = append(filterParts, fmt.Sprintf("rating >= %f", filters.MinRating))
	}

	if len(filterParts) > 0 {
		searchReq.Filter = strings.Join(filterParts, " AND ")
	}

	// Add geo-based sorting/filtering if coordinates are provided.
	if filters.Lat != 0 && filters.Lng != 0 {
		searchReq.Sort = []string{
			fmt.Sprintf("_geoPoint(%f, %f):asc", filters.Lat, filters.Lng),
		}

		if filters.RadiusKM > 0 {
			geoFilter := fmt.Sprintf("_geoRadius(%f, %f, %d)", filters.Lat, filters.Lng, int(filters.RadiusKM*1000))
			if searchReq.Filter != nil {
				searchReq.Filter = fmt.Sprintf("%s AND %s", searchReq.Filter, geoFilter)
			} else {
				searchReq.Filter = geoFilter
			}
		}
	}

	resp, err := index.Search(query, searchReq)
	if err != nil {
		return nil, fmt.Errorf("search providers: %w", err)
	}

	results := &SearchResults{
		Total:            resp.EstimatedTotalHits,
		ProcessingTimeMs: resp.ProcessingTimeMs,
	}

	// Convert hits to ProviderDocuments.
	for _, hit := range resp.Hits {
		var provider ProviderDocument
		if err := hit.DecodeInto(&provider); err != nil {
			continue
		}
		results.Hits = append(results.Hits, provider)
	}

	log.Debug().
		Str("query", query).
		Int64("total", results.Total).
		Int64("processing_ms", results.ProcessingTimeMs).
		Msg("meilisearch provider search completed")

	return results, nil
}

// SearchCategories searches for categories matching the query.
func (s *SearchClient) SearchCategories(_ context.Context, query string) ([]CategoryDocument, error) {
	index := s.client.Index(indexCategories)

	resp, err := index.Search(query, &meilisearch.SearchRequest{
		Limit: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("search categories: %w", err)
	}

	var categories []CategoryDocument
	for _, hit := range resp.Hits {
		var cat CategoryDocument
		if err := hit.DecodeInto(&cat); err != nil {
			continue
		}
		categories = append(categories, cat)
	}

	log.Debug().
		Str("query", query).
		Int("results", len(categories)).
		Msg("meilisearch category search completed")

	return categories, nil
}

// DeleteProvider removes a provider from the search index.
func (s *SearchClient) DeleteProvider(_ context.Context, providerID string) error {
	index := s.client.Index(indexProviders)

	_, err := index.DeleteDocument(providerID, nil)
	if err != nil {
		return fmt.Errorf("search delete provider %s: %w", providerID, err)
	}

	log.Debug().Str("provider_id", providerID).Msg("provider deleted from meilisearch")

	return nil
}
