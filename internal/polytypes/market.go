package polytypes

import "github.com/TrebuchetDynamics/polygolem/pkg/types"

// Gamma protocol DTOs are owned by pkg/types so public SDK signatures do not
// leak internal packages. These aliases preserve the internal call sites.
type HealthResponse = types.HealthResponse
type ImageOptimized = types.ImageOptimized
type Market = types.Market
type Event = types.Event
type Category = types.Category
type Tag = types.Tag
type Series = types.Series
type Collection = types.Collection
type EventCreator = types.EventCreator
type Team = types.Team
type SportMetadata = types.SportMetadata
type TagRelationship = types.TagRelationship
type TagStatus = types.TagStatus
type SearchTag = types.SearchTag
type Profile = types.Profile
type Pagination = types.Pagination
type SearchResponse = types.SearchResponse
type GetMarketsParams = types.GetMarketsParams
type GetEventsParams = types.GetEventsParams
type GetSeriesParams = types.GetSeriesParams
type GetTagsParams = types.GetTagsParams
type SearchParams = types.SearchParams
type GetTeamsParams = types.GetTeamsParams

const (
	TagStatusActive = types.TagStatusActive
	TagStatusClosed = types.TagStatusClosed
	TagStatusAll    = types.TagStatusAll
)
