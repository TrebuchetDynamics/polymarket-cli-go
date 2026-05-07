// Package polytypes — Gamma market and event types stolen from polymarket-go-gamma-client.
package polytypes

// HealthResponse represents the Gamma health check response.
type HealthResponse struct {
	Data string `json:"data"`
}

// ImageOptimized represents an optimized image resource.
type ImageOptimized struct {
	ID                        string         `json:"id"`
	ImageURLSource            string         `json:"imageUrlSource"`
	ImageURLOptimized         string         `json:"imageUrlOptimized"`
	ImageSizeKBSource         int            `json:"imageSizeKbSource"`
	ImageSizeKBOptimized      int            `json:"imageSizeKbOptimized"`
	ImageOptimizedComplete    bool           `json:"imageOptimizedComplete"`
	ImageOptimizedLastUpdated NormalizedTime `json:"imageOptimizedLastUpdated"`
	RelID                     int            `json:"relID"`
	Field                     string         `json:"field"`
	Relname                   string         `json:"relname"`
}

// Market represents a market from the Gamma API.
// Contains extensive metadata beyond what the CLOB API provides.
type Market struct {
	ID                           string          `json:"id"`
	Question                     string          `json:"question"`
	ConditionID                  string          `json:"conditionId"`
	Slug                         string          `json:"slug"`
	QuestionID                   string          `json:"questionID"`
	TwitterCardImage             string          `json:"twitterCardImage"`
	Image                        string          `json:"image"`
	Icon                         string          `json:"icon"`
	Description                  string          `json:"description"`
	ResolutionSource             string          `json:"resolutionSource"`
	EndDate                      NormalizedTime  `json:"endDate"`
	StartDate                    NormalizedTime  `json:"startDate"`
	EndDateISO                   string          `json:"endDateIso"`
	StartDateISO                 string          `json:"startDateIso"`
	UMAEndDate                   NormalizedTime  `json:"umaEndDate"`
	UMAEndDateISO                string          `json:"umaEndDateIso"`
	ClosedTime                   NormalizedTime  `json:"closedTime"`
	Category                     string          `json:"category"`
	AmmType                      string          `json:"ammType"`
	Liquidity                    string          `json:"liquidity"`
	LiquidityNum                 float64         `json:"liquidityNum"`
	Volume                       string          `json:"volume"`
	VolumeNum                    float64         `json:"volumeNum"`
	Fee                          string          `json:"fee"`
	DenominationToken            string          `json:"denominationToken"`
	SponsorName                  string          `json:"sponsorName"`
	SponsorImage                 string          `json:"sponsorImage"`
	XAxisValue                   string          `json:"xAxisValue"`
	YAxisValue                   string          `json:"yAxisValue"`
	LowerBound                   string          `json:"lowerBound"`
	UpperBound                   string          `json:"upperBound"`
	Outcomes                     StringOrArray   `json:"outcomes"`
	OutcomePrices                StringOrArray   `json:"outcomePrices"`
	ShortOutcomes                StringOrArray   `json:"shortOutcomes"`
	Active                       bool            `json:"active"`
	Closed                       bool            `json:"closed"`
	Archived                     bool            `json:"archived"`
	New                          bool            `json:"new"`
	Featured                     bool            `json:"featured"`
	Restricted                   bool            `json:"restricted"`
	WideFormat                   bool            `json:"wideFormat"`
	Ready                        bool            `json:"ready"`
	Funded                       bool            `json:"funded"`
	MarketType                   string          `json:"marketType"`
	FormatType                   string          `json:"formatType"`
	LowerBoundDate               NormalizedTime  `json:"lowerBoundDate"`
	UpperBoundDate               NormalizedTime  `json:"upperBoundDate"`
	MarketMakerAddress           string          `json:"marketMakerAddress"`
	CreatedBy                    int             `json:"createdBy"`
	UpdatedBy                    int             `json:"updatedBy"`
	CreatedAt                    NormalizedTime  `json:"createdAt"`
	UpdatedAt                    NormalizedTime  `json:"updatedAt"`
	MailchimpTag                 string          `json:"mailchimpTag"`
	ResolvedBy                   string          `json:"resolvedBy"`
	MarketGroup                  int             `json:"marketGroup"`
	GroupItemTitle               string          `json:"groupItemTitle"`
	GroupItemThreshold           string          `json:"groupItemThreshold"`
	GroupItemRange               string          `json:"groupItemRange"`
	UMAResolutionStatus          string          `json:"umaResolutionStatus"`
	UMAResolutionStatuses        string          `json:"umaResolutionStatuses"`
	UMABond                      string          `json:"umaBond"`
	UMAReward                    string          `json:"umaReward"`
	EnableOrderBook              bool            `json:"enableOrderBook"`
	OrderPriceMinTickSize        float64         `json:"orderPriceMinTickSize"`
	OrderMinSize                 float64         `json:"orderMinSize"`
	MakerBaseFee                 int             `json:"makerBaseFee"`
	TakerBaseFee                 int             `json:"takerBaseFee"`
	AcceptingOrders              bool            `json:"acceptingOrders"`
	NotificationsEnabled         bool            `json:"notificationsEnabled"`
	CurationOrder                int             `json:"curationOrder"`
	Score                        float64         `json:"score"`
	HasReviewedDates             bool            `json:"hasReviewedDates"`
	ReadyForCron                 bool            `json:"readyForCron"`
	CommentsEnabled              bool            `json:"commentsEnabled"`
	Volume24hr                   float64         `json:"volume24hr"`
	Volume1wk                    float64         `json:"volume1wk"`
	Volume1mo                    float64         `json:"volume1mo"`
	Volume1yr                    float64         `json:"volume1yr"`
	Volume24hrAmm                float64         `json:"volume24hrAmm"`
	Volume1wkAmm                 float64         `json:"volume1wkAmm"`
	Volume1moAmm                 float64         `json:"volume1moAmm"`
	Volume1yrAmm                 float64         `json:"volume1yrAmm"`
	Volume24hrClob               float64         `json:"volume24hrClob"`
	Volume1wkClob                float64         `json:"volume1wkClob"`
	Volume1moClob                float64         `json:"volume1moClob"`
	Volume1yrClob                float64         `json:"volume1yrClob"`
	VolumeAmm                    float64         `json:"volumeAmm"`
	VolumeClob                   float64         `json:"volumeClob"`
	LiquidityAmm                 float64         `json:"liquidityAmm"`
	LiquidityClob                float64         `json:"liquidityClob"`
	GameStartTime                NormalizedTime  `json:"gameStartTime"`
	SecondsDelay                 int             `json:"secondsDelay"`
	ClobTokenIDs                 string          `json:"clobTokenIds"`
	TeamAID                      string          `json:"teamAID"`
	TeamBID                      string          `json:"teamBID"`
	GameID                       string          `json:"gameId"`
	SportsMarketType             string          `json:"sportsMarketType"`
	Line                         float64         `json:"line"`
	DisqusThread                 string          `json:"disqusThread"`
	FPMMLive                     bool            `json:"fpmmLive"`
	CustomLiveness               int             `json:"customLiveness"`
	RewardsMinSize               float64         `json:"rewardsMinSize"`
	RewardsMaxSpread             float64         `json:"rewardsMaxSpread"`
	ImageOptimized               *ImageOptimized `json:"imageOptimized,omitempty"`
	IconOptimized                *ImageOptimized `json:"iconOptimized,omitempty"`
	Events                       []Event         `json:"events,omitempty"`
	Categories                   []Category      `json:"categories,omitempty"`
	Tags                         []Tag           `json:"tags,omitempty"`
	Creator                      string          `json:"creator"`
	PastSlugs                    string          `json:"pastSlugs"`
	ReadyTimestamp               NormalizedTime  `json:"readyTimestamp"`
	FundedTimestamp              NormalizedTime  `json:"fundedTimestamp"`
	AcceptingOrdersTimestamp     NormalizedTime  `json:"acceptingOrdersTimestamp"`
	Competitive                  float64         `json:"competitive"`
	Spread                       float64         `json:"spread"`
	AutomaticallyResolved        bool            `json:"automaticallyResolved"`
	OneDayPriceChange            float64         `json:"oneDayPriceChange"`
	OneHourPriceChange           float64         `json:"oneHourPriceChange"`
	OneWeekPriceChange           float64         `json:"oneWeekPriceChange"`
	OneMonthPriceChange          float64         `json:"oneMonthPriceChange"`
	OneYearPriceChange           float64         `json:"oneYearPriceChange"`
	LastTradePrice               float64         `json:"lastTradePrice"`
	BestBid                      float64         `json:"bestBid"`
	BestAsk                      float64         `json:"bestAsk"`
	AutomaticallyActive          bool            `json:"automaticallyActive"`
	ClearBookOnStart             bool            `json:"clearBookOnStart"`
	ManualActivation             bool            `json:"manualActivation"`
	ChartColor                   string          `json:"chartColor"`
	SeriesColor                  string          `json:"seriesColor"`
	ShowGmpSeries                bool            `json:"showGmpSeries"`
	ShowGmpOutcome               bool            `json:"showGmpOutcome"`
	NegRiskOther                 bool            `json:"negRiskOther"`
	PendingDeployment            bool            `json:"pendingDeployment"`
	Deploying                    bool            `json:"deploying"`
	DeployingTimestamp           NormalizedTime  `json:"deployingTimestamp"`
	ScheduledDeploymentTimestamp NormalizedTime  `json:"scheduledDeploymentTimestamp"`
	RFQEnabled                   bool            `json:"rfqEnabled"`
	EventStartTime               NormalizedTime  `json:"eventStartTime"`
}

// Event represents an event from the Gamma API.
type Event struct {
	ID                           string          `json:"id"`
	Ticker                       string          `json:"ticker"`
	Slug                         string          `json:"slug"`
	Title                        string          `json:"title"`
	Subtitle                     string          `json:"subtitle"`
	Description                  string          `json:"description"`
	ResolutionSource             string          `json:"resolutionSource"`
	StartDate                    NormalizedTime  `json:"startDate"`
	CreationDate                 NormalizedTime  `json:"creationDate"`
	EndDate                      NormalizedTime  `json:"endDate"`
	Image                        string          `json:"image"`
	Icon                         string          `json:"icon"`
	Active                       bool            `json:"active"`
	Closed                       bool            `json:"closed"`
	Archived                     bool            `json:"archived"`
	New                          bool            `json:"new"`
	Featured                     bool            `json:"featured"`
	Restricted                   bool            `json:"restricted"`
	Liquidity                    float64         `json:"liquidity"`
	Volume                       float64         `json:"volume"`
	OpenInterest                 float64         `json:"openInterest"`
	SortBy                       string          `json:"sortBy"`
	Category                     string          `json:"category"`
	Subcategory                  string          `json:"subcategory"`
	IsTemplate                   bool            `json:"isTemplate"`
	TemplateVariables            string          `json:"templateVariables"`
	PublishedAt                  NormalizedTime  `json:"published_at"`
	CreatedBy                    string          `json:"createdBy"`
	UpdatedBy                    string          `json:"updatedBy"`
	CreatedAt                    NormalizedTime  `json:"createdAt"`
	UpdatedAt                    NormalizedTime  `json:"updatedAt"`
	CommentsEnabled              bool            `json:"commentsEnabled"`
	Competitive                  float64         `json:"competitive"`
	Volume24hr                   float64         `json:"volume24hr"`
	Volume1wk                    float64         `json:"volume1wk"`
	Volume1mo                    float64         `json:"volume1mo"`
	Volume1yr                    float64         `json:"volume1yr"`
	FeaturedImage                string          `json:"featuredImage"`
	DisqusThread                 string          `json:"disqusThread"`
	ParentEvent                  string          `json:"parentEvent"`
	EnableOrderBook              bool            `json:"enableOrderBook"`
	LiquidityAmm                 float64         `json:"liquidityAmm"`
	LiquidityClob                float64         `json:"liquidityClob"`
	NegRisk                      bool            `json:"negRisk"`
	NegRiskMarketID              string          `json:"negRiskMarketID"`
	NegRiskFeeBips               int             `json:"negRiskFeeBips"`
	CommentCount                 int             `json:"commentCount"`
	ImageOptimized               *ImageOptimized `json:"imageOptimized,omitempty"`
	IconOptimized                *ImageOptimized `json:"iconOptimized,omitempty"`
	FeaturedImageOptimized       *ImageOptimized `json:"featuredImageOptimized,omitempty"`
	SubEvents                    []string        `json:"subEvents,omitempty"`
	Markets                      []Market        `json:"markets,omitempty"`
	Series                       []Series        `json:"series,omitempty"`
	Categories                   []Category      `json:"categories,omitempty"`
	Collections                  []Collection    `json:"collections,omitempty"`
	Tags                         []Tag           `json:"tags,omitempty"`
	CYOM                         bool            `json:"cyom"`
	ClosedTime                   NormalizedTime  `json:"closedTime"`
	ShowAllOutcomes              bool            `json:"showAllOutcomes"`
	ShowMarketImages             bool            `json:"showMarketImages"`
	AutomaticallyResolved        bool            `json:"automaticallyResolved"`
	EnableNegRisk                bool            `json:"enableNegRisk"`
	AutomaticallyActive          bool            `json:"automaticallyActive"`
	EventDate                    NormalizedTime  `json:"eventDate"`
	StartTime                    NormalizedTime  `json:"startTime"`
	EventWeek                    int             `json:"eventWeek"`
	SeriesSlug                   string          `json:"seriesSlug"`
	Score                        string          `json:"score"`
	Elapsed                      string          `json:"elapsed"`
	Period                       string          `json:"period"`
	Live                         bool            `json:"live"`
	Ended                        bool            `json:"ended"`
	FinishedTimestamp            NormalizedTime  `json:"finishedTimestamp"`
	GMPChartMode                 string          `json:"gmpChartMode"`
	EventCreators                []EventCreator  `json:"eventCreators,omitempty"`
	TweetCount                   int             `json:"tweetCount"`
	FeaturedOrder                int             `json:"featuredOrder"`
	EstimateValue                bool            `json:"estimateValue"`
	CantEstimate                 bool            `json:"cantEstimate"`
	EstimatedValue               string          `json:"estimatedValue"`
	SpreadsMainLine              float64         `json:"spreadsMainLine"`
	TotalsMainLine               float64         `json:"totalsMainLine"`
	CarouselMap                  string          `json:"carouselMap"`
	PendingDeployment            bool            `json:"pendingDeployment"`
	Deploying                    bool            `json:"deploying"`
	DeployingTimestamp           NormalizedTime  `json:"deployingTimestamp"`
	ScheduledDeploymentTimestamp NormalizedTime  `json:"scheduledDeploymentTimestamp"`
	GameStatus                   string          `json:"gameStatus"`
}

// Category represents a market category.
type Category struct {
	ID             string         `json:"id"`
	Label          string         `json:"label"`
	ParentCategory string         `json:"parentCategory"`
	Slug           string         `json:"slug"`
	PublishedAt    NormalizedTime `json:"publishedAt"`
	CreatedBy      string         `json:"createdBy"`
	UpdatedBy      string         `json:"updatedBy"`
	CreatedAt      NormalizedTime `json:"createdAt"`
	UpdatedAt      NormalizedTime `json:"updatedAt"`
}

// Tag represents a tag associated with markets or events.
type Tag struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`
	Slug        string         `json:"slug"`
	ForceShow   bool           `json:"forceShow"`
	PublishedAt NormalizedTime `json:"publishedAt"`
	CreatedBy   int            `json:"createdBy"`
	UpdatedBy   int            `json:"updatedBy"`
	CreatedAt   NormalizedTime `json:"createdAt"`
	UpdatedAt   NormalizedTime `json:"updatedAt"`
	ForceHide   bool           `json:"forceHide"`
	IsCarousel  bool           `json:"isCarousel"`
}

// Series represents a series of events.
type Series struct {
	ID                string         `json:"id"`
	Ticker            string         `json:"ticker"`
	Slug              string         `json:"slug"`
	Title             string         `json:"title"`
	Subtitle          string         `json:"subtitle"`
	SeriesType        string         `json:"seriesType"`
	Recurrence        string         `json:"recurrence"`
	Description       string         `json:"description"`
	Image             string         `json:"image"`
	Icon              string         `json:"icon"`
	Layout            string         `json:"layout"`
	Active            bool           `json:"active"`
	Closed            bool           `json:"closed"`
	Archived          bool           `json:"archived"`
	New               bool           `json:"new"`
	Featured          bool           `json:"featured"`
	Restricted        bool           `json:"restricted"`
	IsTemplate        bool           `json:"isTemplate"`
	TemplateVariables bool           `json:"templateVariables"`
	PublishedAt       NormalizedTime `json:"publishedAt"`
	CreatedBy         string         `json:"createdBy"`
	UpdatedBy         string         `json:"updatedBy"`
	CreatedAt         NormalizedTime `json:"createdAt"`
	UpdatedAt         NormalizedTime `json:"updatedAt"`
	CommentsEnabled   bool           `json:"commentsEnabled"`
	Competitive       string         `json:"competitive"`
	Volume24hr        float64        `json:"volume24hr"`
	Volume            float64        `json:"volume"`
	Liquidity         float64        `json:"liquidity"`
	StartDate         NormalizedTime `json:"startDate"`
	PythTokenID       string         `json:"pythTokenID"`
	CGAssetName       string         `json:"cgAssetName"`
	Score             float64        `json:"score"`
	Events            []Event        `json:"events,omitempty"`
	Collections       []Collection   `json:"collections,omitempty"`
	Categories        []Category     `json:"categories,omitempty"`
	Tags              []Tag          `json:"tags,omitempty"`
	CommentCount      int            `json:"commentCount"`
}

// Collection represents a collection of events.
type Collection struct {
	ID                   string          `json:"id"`
	Ticker               string          `json:"ticker"`
	Slug                 string          `json:"slug"`
	Title                string          `json:"title"`
	Subtitle             string          `json:"subtitle"`
	CollectionType       string          `json:"collectionType"`
	Description          string          `json:"description"`
	Tags                 string          `json:"tags"`
	Image                string          `json:"image"`
	Icon                 string          `json:"icon"`
	HeaderImage          string          `json:"headerImage"`
	Layout               string          `json:"layout"`
	Active               bool            `json:"active"`
	Closed               bool            `json:"closed"`
	Archived             bool            `json:"archived"`
	New                  bool            `json:"new"`
	Featured             bool            `json:"featured"`
	Restricted           bool            `json:"restricted"`
	IsTemplate           bool            `json:"isTemplate"`
	TemplateVariables    string          `json:"templateVariables"`
	PublishedAt          NormalizedTime  `json:"publishedAt"`
	CreatedBy            string          `json:"createdBy"`
	UpdatedBy            string          `json:"updatedBy"`
	CreatedAt            NormalizedTime  `json:"createdAt"`
	UpdatedAt            NormalizedTime  `json:"updatedAt"`
	CommentsEnabled      bool            `json:"commentsEnabled"`
	ImageOptimized       *ImageOptimized `json:"imageOptimized,omitempty"`
	IconOptimized        *ImageOptimized `json:"iconOptimized,omitempty"`
	HeaderImageOptimized *ImageOptimized `json:"headerImageOptimized,omitempty"`
}

// EventCreator represents a creator of an event.
type EventCreator struct {
	ID            string         `json:"id"`
	CreatorName   string         `json:"creatorName"`
	CreatorHandle string         `json:"creatorHandle"`
	CreatorURL    string         `json:"creatorUrl"`
	CreatorImage  string         `json:"creatorImage"`
	CreatedAt     NormalizedTime `json:"createdAt"`
	UpdatedAt     NormalizedTime `json:"updatedAt"`
}

// Team represents a sports team.
type Team struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	League       string         `json:"league"`
	Record       string         `json:"record"`
	Logo         string         `json:"logo"`
	Abbreviation string         `json:"abbreviation"`
	Alias        string         `json:"alias"`
	CreatedAt    NormalizedTime `json:"createdAt"`
	UpdatedAt    NormalizedTime `json:"updatedAt"`
}

// SportMetadata represents metadata for a sport.
type SportMetadata struct {
	Sport      string `json:"sport"`
	Image      string `json:"image"`
	Resolution string `json:"resolution"`
	Ordering   string `json:"ordering"`
	Tags       string `json:"tags"`
	Series     string `json:"series"`
}

// TagRelationship represents a relationship between two tags.
type TagRelationship struct {
	ID           string `json:"id"`
	TagID        int    `json:"tagID"`
	RelatedTagID int    `json:"relatedTagID"`
	Rank         int    `json:"rank"`
}

// TagStatus represents the status filter for related tags.
type TagStatus string

const (
	TagStatusActive TagStatus = "active"
	TagStatusClosed TagStatus = "closed"
	TagStatusAll    TagStatus = "all"
)

// SearchTag represents a tag in search results.
type SearchTag struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Slug       string `json:"slug"`
	EventCount int    `json:"event_count"`
}

// Profile represents a user profile.
type Profile struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	User                  int             `json:"user"`
	Referral              string          `json:"referral"`
	CreatedBy             int             `json:"createdBy"`
	UpdatedBy             int             `json:"updatedBy"`
	CreatedAt             NormalizedTime  `json:"createdAt"`
	UpdatedAt             NormalizedTime  `json:"updatedAt"`
	UTMSource             string          `json:"utmSource"`
	UTMMedium             string          `json:"utmMedium"`
	UTMCampaign           string          `json:"utmCampaign"`
	UTMContent            string          `json:"utmContent"`
	UTMTerm               string          `json:"utmTerm"`
	WalletActivated       bool            `json:"walletActivated"`
	Pseudonym             string          `json:"pseudonym"`
	DisplayUsernamePublic bool            `json:"displayUsernamePublic"`
	ProfileImage          string          `json:"profileImage"`
	Bio                   string          `json:"bio"`
	ProxyWallet           string          `json:"proxyWallet"`
	ProfileImageOptimized *ImageOptimized `json:"profileImageOptimized,omitempty"`
	IsCloseOnly           bool            `json:"isCloseOnly"`
	IsCertReq             bool            `json:"isCertReq"`
	CertReqDate           NormalizedTime  `json:"certReqDate"`
}

// Pagination represents pagination info.
type Pagination struct {
	HasMore      bool `json:"hasMore"`
	TotalResults int  `json:"totalResults"`
}

// SearchResponse represents the Gamma search response.
type SearchResponse struct {
	Events     []Event     `json:"events"`
	Tags       []SearchTag `json:"tags"`
	Profiles   []Profile   `json:"profiles"`
	Pagination Pagination  `json:"pagination"`
}

// --- Gamma Query Params (stolen from polymarket-go-gamma-client) ---

type GetMarketsParams struct {
	Limit             int             `json:"limit,omitempty"`
	Offset            int             `json:"offset,omitempty"`
	Order             string          `json:"order,omitempty"`
	Ascending         *bool           `json:"ascending,omitempty"`
	ID                []int           `json:"id,omitempty"`
	Slug              []string        `json:"slug,omitempty"`
	ClobTokenIDs      []string        `json:"clob_token_ids,omitempty"`
	ConditionIDs      []string        `json:"condition_ids,omitempty"`
	TagID             *int            `json:"tag_id,omitempty"`
	RelatedTags       *bool           `json:"related_tags,omitempty"`
	Closed            *bool           `json:"closed,omitempty"`
	Active            *bool           `json:"active,omitempty"`
	LiquidityNumMin   *float64        `json:"liquidity_num_min,omitempty"`
	LiquidityNumMax   *float64        `json:"liquidity_num_max,omitempty"`
	VolumeNumMin      *float64        `json:"volume_num_min,omitempty"`
	VolumeNumMax      *float64        `json:"volume_num_max,omitempty"`
	StartDateMin      *NormalizedTime `json:"start_date_min,omitempty"`
	StartDateMax      *NormalizedTime `json:"start_date_max,omitempty"`
	EndDateMin        *NormalizedTime `json:"end_date_min,omitempty"`
	EndDateMax        *NormalizedTime `json:"end_date_max,omitempty"`
	RewardsMinSize    *float64        `json:"rewards_min_size,omitempty"`
	SportsMarketTypes []string        `json:"sports_market_types,omitempty"`
	GameID            string          `json:"game_id,omitempty"`
}

type GetEventsParams struct {
	Limit        int             `json:"limit,omitempty"`
	Offset       int             `json:"offset,omitempty"`
	Order        string          `json:"order,omitempty"`
	Ascending    *bool           `json:"ascending,omitempty"`
	ID           []int           `json:"id,omitempty"`
	Slug         []string        `json:"slug,omitempty"`
	TagID        *int            `json:"tag_id,omitempty"`
	RelatedTags  *bool           `json:"related_tags,omitempty"`
	Featured     *bool           `json:"featured,omitempty"`
	Closed       *bool           `json:"closed,omitempty"`
	StartDateMin *NormalizedTime `json:"start_date_min,omitempty"`
	StartDateMax *NormalizedTime `json:"start_date_max,omitempty"`
	EndDateMin   *NormalizedTime `json:"end_date_min,omitempty"`
	EndDateMax   *NormalizedTime `json:"end_date_max,omitempty"`
	Recurrence   string          `json:"recurrence,omitempty"`
}

type GetSeriesParams struct {
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
	Order     string   `json:"order,omitempty"`
	Ascending *bool    `json:"ascending,omitempty"`
	Slug      []string `json:"slug,omitempty"`
	Closed    *bool    `json:"closed,omitempty"`
}

type GetTagsParams struct {
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	Order     string `json:"order,omitempty"`
	Ascending *bool  `json:"ascending,omitempty"`
}

type SearchParams struct {
	Q              string   `json:"q"`
	LimitPerType   *int     `json:"limit_per_type,omitempty"`
	Page           *int     `json:"page,omitempty"`
	EventsTag      []string `json:"events_tag,omitempty"`
	EventsStatus   string   `json:"events_status,omitempty"`
	Ascending      *bool    `json:"ascending,omitempty"`
	Sort           string   `json:"sort,omitempty"`
	SearchProfiles *bool    `json:"search_profiles,omitempty"`
}

type GetTeamsParams struct {
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
	Order        string   `json:"order,omitempty"`
	Ascending    *bool    `json:"ascending,omitempty"`
	League       []string `json:"league,omitempty"`
	Name         []string `json:"name,omitempty"`
	Abbreviation []string `json:"abbreviation,omitempty"`
}
