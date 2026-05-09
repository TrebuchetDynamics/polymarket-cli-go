// Package orderresults joins Polymarket account history into one read-only
// operator report: Data API positions/results plus optional authenticated CLOB
// open orders and trade history.
package orderresults

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdkclob "github.com/TrebuchetDynamics/polygolem/pkg/clob"
	"github.com/TrebuchetDynamics/polygolem/pkg/types"
)

const (
	StatusOpen    = "open"
	StatusWon     = "won"
	StatusLost    = "lost"
	StatusClosed  = "closed"
	StatusUnknown = "unknown"

	SourceData = "data"
	SourceCLOB = "clob"
)

// DataReader is the minimal read-only Data API contract used by BuildReport.
type DataReader interface {
	CurrentPositionsWithLimit(context.Context, string, int) ([]types.Position, error)
	ClosedPositionsWithLimit(context.Context, string, int) ([]types.ClosedPosition, error)
	Trades(context.Context, string, int) ([]types.Trade, error)
}

// CLOBReader is the optional authenticated CLOB account-history contract used
// when Options.IncludeCLOB is set.
type CLOBReader interface {
	ListOrders(context.Context, string) ([]sdkclob.OrderRecord, error)
	ListTrades(context.Context, string) ([]sdkclob.TradeRecord, error)
}

// Source adapts separate public SDK clients into one BuildReport source.
type Source struct {
	Data DataReader
	CLOB CLOBReader
}

func (s Source) CurrentPositionsWithLimit(ctx context.Context, user string, limit int) ([]types.Position, error) {
	if s.Data == nil {
		return nil, fmt.Errorf("orderresults: data reader is required")
	}
	return s.Data.CurrentPositionsWithLimit(ctx, user, limit)
}

func (s Source) ClosedPositionsWithLimit(ctx context.Context, user string, limit int) ([]types.ClosedPosition, error) {
	if s.Data == nil {
		return nil, fmt.Errorf("orderresults: data reader is required")
	}
	return s.Data.ClosedPositionsWithLimit(ctx, user, limit)
}

func (s Source) Trades(ctx context.Context, user string, limit int) ([]types.Trade, error) {
	if s.Data == nil {
		return nil, fmt.Errorf("orderresults: data reader is required")
	}
	return s.Data.Trades(ctx, user, limit)
}

func (s Source) ListOrders(ctx context.Context, privateKey string) ([]sdkclob.OrderRecord, error) {
	if s.CLOB == nil {
		return nil, fmt.Errorf("orderresults: clob reader is required")
	}
	return s.CLOB.ListOrders(ctx, privateKey)
}

func (s Source) ListTrades(ctx context.Context, privateKey string) ([]sdkclob.TradeRecord, error) {
	if s.CLOB == nil {
		return nil, fmt.Errorf("orderresults: clob reader is required")
	}
	return s.CLOB.ListTrades(ctx, privateKey)
}

type Options struct {
	Limit       int
	IncludeCLOB bool
	PrivateKey  string
}

type Report struct {
	User     string   `json:"user"`
	Limit    int      `json:"limit"`
	Summary  Summary  `json:"summary"`
	Rows     []Row    `json:"rows"`
	Warnings []string `json:"warnings,omitempty"`
}

type Summary struct {
	Positions       int     `json:"positions"`
	ClosedPositions int     `json:"closedPositions"`
	Closed          int     `json:"closed"`
	DataTrades      int     `json:"dataTrades"`
	CLOBTrades      int     `json:"clobTrades"`
	OpenOrders      int     `json:"openOrders"`
	Redeemable      int     `json:"redeemable"`
	Won             int     `json:"won"`
	Lost            int     `json:"lost"`
	Open            int     `json:"open"`
	Unknown         int     `json:"unknown"`
	InitialValue    float64 `json:"initialValue"`
	CurrentValue    float64 `json:"currentValue"`
	CashPnl         float64 `json:"cashPnl"`
	RealizedPnl     float64 `json:"realizedPnl"`
	MatchedNotional float64 `json:"matchedNotional"`
}

type Row struct {
	Market         string         `json:"market,omitempty"`
	TokenID        string         `json:"tokenId,omitempty"`
	Title          string         `json:"title,omitempty"`
	Slug           string         `json:"slug,omitempty"`
	Outcome        string         `json:"outcome,omitempty"`
	Status         string         `json:"status"`
	Redeemable     bool           `json:"redeemable,omitempty"`
	Mergeable      bool           `json:"mergeable,omitempty"`
	NegativeRisk   bool           `json:"negativeRisk,omitempty"`
	Size           float64        `json:"size,omitempty"`
	AvgPrice       float64        `json:"avgPrice,omitempty"`
	InitialValue   float64        `json:"initialValue,omitempty"`
	CurrentPrice   float64        `json:"curPrice,omitempty"`
	CurrentValue   float64        `json:"currentValue,omitempty"`
	CashPnl        float64        `json:"cashPnl,omitempty"`
	PercentPnl     float64        `json:"percentPnl,omitempty"`
	RealizedPnl    float64        `json:"realizedPnl,omitempty"`
	EndDate        string         `json:"endDate,omitempty"`
	TradeCount     int            `json:"tradeCount"`
	OpenOrderCount int            `json:"openOrderCount"`
	Trades         []TradeSummary `json:"trades,omitempty"`
	OpenOrders     []OrderSummary `json:"openOrders,omitempty"`
}

type TradeSummary struct {
	Source          string  `json:"source"`
	ID              string  `json:"id,omitempty"`
	Status          string  `json:"status,omitempty"`
	Side            string  `json:"side,omitempty"`
	Price           float64 `json:"price,omitempty"`
	Size            float64 `json:"size,omitempty"`
	Outcome         string  `json:"outcome,omitempty"`
	Timestamp       string  `json:"timestamp,omitempty"`
	TransactionHash string  `json:"transactionHash,omitempty"`
}

type OrderSummary struct {
	ID           string  `json:"id,omitempty"`
	Status       string  `json:"status,omitempty"`
	Side         string  `json:"side,omitempty"`
	Price        float64 `json:"price,omitempty"`
	OriginalSize float64 `json:"originalSize,omitempty"`
	SizeMatched  float64 `json:"sizeMatched,omitempty"`
	Outcome      string  `json:"outcome,omitempty"`
	CreatedAt    string  `json:"createdAt,omitempty"`
}

func BuildReport(ctx context.Context, source DataReader, user string, opts Options) (*Report, error) {
	user = strings.TrimSpace(user)
	if user == "" {
		return nil, fmt.Errorf("orderresults: user is required")
	}
	if source == nil {
		return nil, fmt.Errorf("orderresults: data reader is required")
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	builder := newReportBuilder(user, limit)

	positions, err := source.CurrentPositionsWithLimit(ctx, user, limit)
	if err != nil {
		return nil, fmt.Errorf("orderresults: positions: %w", err)
	}
	for _, position := range positions {
		if emptyPosition(position) {
			continue
		}
		builder.addPosition(position)
	}

	closed, err := source.ClosedPositionsWithLimit(ctx, user, limit)
	if err != nil {
		return nil, fmt.Errorf("orderresults: closed positions: %w", err)
	}
	for _, position := range closed {
		if emptyClosedPosition(position) {
			continue
		}
		builder.addClosedPosition(position)
	}

	trades, err := source.Trades(ctx, user, limit)
	if err != nil {
		return nil, fmt.Errorf("orderresults: data trades: %w", err)
	}
	for _, trade := range trades {
		if emptyDataTrade(trade) {
			continue
		}
		builder.addDataTrade(trade)
	}

	if opts.IncludeCLOB {
		privateKey := strings.TrimSpace(opts.PrivateKey)
		if privateKey == "" {
			return nil, fmt.Errorf("orderresults: private key is required when IncludeCLOB is true")
		}
		clobSource, ok := source.(CLOBReader)
		if !ok {
			return nil, fmt.Errorf("orderresults: clob reader is required when IncludeCLOB is true")
		}
		orders, err := clobSource.ListOrders(ctx, privateKey)
		if err != nil {
			return nil, fmt.Errorf("orderresults: clob orders: %w", err)
		}
		for _, order := range orders {
			if emptyCLOBOrder(order) {
				continue
			}
			builder.addCLOBOrder(order)
		}
		clobTrades, err := clobSource.ListTrades(ctx, privateKey)
		if err != nil {
			return nil, fmt.Errorf("orderresults: clob trades: %w", err)
		}
		for _, trade := range clobTrades {
			if emptyCLOBTrade(trade) {
				continue
			}
			builder.addCLOBTrade(trade)
		}
	}

	return builder.build(), nil
}

func (r *Report) RowByToken(tokenID string) *Row {
	for i := range r.Rows {
		if r.Rows[i].TokenID == tokenID {
			return &r.Rows[i]
		}
	}
	return nil
}

type reportBuilder struct {
	report Report
	rows   map[string]*Row
	order  []string
}

func newReportBuilder(user string, limit int) *reportBuilder {
	return &reportBuilder{
		report: Report{User: user, Limit: limit},
		rows:   map[string]*Row{},
	}
}

func (b *reportBuilder) addPosition(position types.Position) {
	row := b.row(position.ConditionID, position.TokenID)
	row.Market = firstNonEmpty(row.Market, position.ConditionID)
	row.TokenID = firstNonEmpty(row.TokenID, position.TokenID)
	row.Title = firstNonEmpty(row.Title, position.Title)
	row.Slug = firstNonEmpty(row.Slug, position.Slug)
	row.Outcome = firstNonEmpty(row.Outcome, position.Outcome)
	row.Status = classifyPosition(position)
	row.Redeemable = position.Redeemable
	row.Mergeable = position.Mergeable
	row.NegativeRisk = position.NegativeRisk
	row.Size = position.Size
	row.AvgPrice = position.AvgPrice
	row.InitialValue = position.InitialValue
	row.CurrentPrice = position.CurrentPrice
	row.CurrentValue = position.CurrentValue
	row.CashPnl = position.CashPnl
	row.PercentPnl = position.PercentPnl
	row.RealizedPnl = position.RealizedPnl
	row.EndDate = firstNonEmpty(row.EndDate, position.EndDate)

	b.report.Summary.Positions++
	b.report.Summary.InitialValue += position.InitialValue
	b.report.Summary.CurrentValue += position.CurrentValue
	b.report.Summary.CashPnl += position.CashPnl
	if position.Redeemable {
		b.report.Summary.Redeemable++
	}
	b.countStatus(row.Status)
}

func (b *reportBuilder) addClosedPosition(position types.ClosedPosition) {
	row := b.row(position.ConditionID, position.TokenID)
	row.Market = firstNonEmpty(row.Market, position.ConditionID, position.MarketID)
	row.TokenID = firstNonEmpty(row.TokenID, position.TokenID)
	row.Title = firstNonEmpty(row.Title, position.Title)
	row.Slug = firstNonEmpty(row.Slug, position.Slug)
	row.Outcome = firstNonEmpty(row.Outcome, position.Outcome, position.Side)
	if row.Status == "" || row.Status == StatusUnknown {
		row.Status = StatusClosed
	}
	row.Size = firstNonZero(row.Size, position.Size, position.TotalBought)
	row.AvgPrice = firstNonZero(row.AvgPrice, position.AvgPrice, position.AvgPriceBuy)
	row.CurrentPrice = firstNonZero(row.CurrentPrice, position.CurrentPrice)
	row.RealizedPnl += position.RealizedPnl
	row.EndDate = firstNonEmpty(row.EndDate, position.EndDate)

	b.report.Summary.ClosedPositions++
	b.report.Summary.RealizedPnl += position.RealizedPnl
	if row.Status == StatusClosed {
		b.countStatus(StatusClosed)
	}
}

func (b *reportBuilder) addDataTrade(trade types.Trade) {
	row := b.row(trade.Market, trade.AssetID)
	row.Market = firstNonEmpty(row.Market, trade.Market)
	row.TokenID = firstNonEmpty(row.TokenID, trade.AssetID)
	row.Title = firstNonEmpty(row.Title, trade.Title)
	row.Slug = firstNonEmpty(row.Slug, trade.Slug)
	row.Outcome = firstNonEmpty(row.Outcome, trade.Outcome)
	if row.Status == "" {
		row.Status = StatusUnknown
	}
	added := appendTrade(row, TradeSummary{
		Source:          SourceData,
		ID:              trade.ID,
		Status:          trade.Status,
		Side:            trade.Side,
		Price:           trade.Price,
		Size:            trade.Size,
		Outcome:         trade.Outcome,
		Timestamp:       trade.CreatedAt,
		TransactionHash: trade.TransactionHash,
	})
	row.TradeCount = len(row.Trades)
	b.report.Summary.DataTrades++
	if added {
		b.report.Summary.MatchedNotional += trade.Price * trade.Size
	}
}

func (b *reportBuilder) addCLOBOrder(order sdkclob.OrderRecord) {
	row := b.row(order.Market, order.AssetID)
	row.Market = firstNonEmpty(row.Market, order.Market)
	row.TokenID = firstNonEmpty(row.TokenID, order.AssetID)
	row.Outcome = firstNonEmpty(row.Outcome, order.Outcome)
	if row.Status == "" || row.Status == StatusUnknown {
		row.Status = StatusOpen
	}
	row.OpenOrders = append(row.OpenOrders, OrderSummary{
		ID:           order.ID,
		Status:       order.Status,
		Side:         order.Side,
		Price:        parseFloat(order.Price),
		OriginalSize: parseFloat(order.OriginalSize),
		SizeMatched:  parseFloat(order.SizeMatched),
		Outcome:      order.Outcome,
		CreatedAt:    order.CreatedAt,
	})
	row.OpenOrderCount = len(row.OpenOrders)
	b.report.Summary.OpenOrders++
}

func (b *reportBuilder) addCLOBTrade(trade sdkclob.TradeRecord) {
	row := b.row(trade.Market, trade.AssetID)
	row.Market = firstNonEmpty(row.Market, trade.Market)
	row.TokenID = firstNonEmpty(row.TokenID, trade.AssetID)
	row.Outcome = firstNonEmpty(row.Outcome, trade.Outcome)
	if row.Status == "" {
		row.Status = StatusUnknown
	}
	price := parseFloat(trade.Price)
	size := parseFloat(trade.Size)
	added := appendTrade(row, TradeSummary{
		Source:          SourceCLOB,
		ID:              trade.ID,
		Status:          trade.Status,
		Side:            trade.Side,
		Price:           price,
		Size:            size,
		Outcome:         trade.Outcome,
		Timestamp:       firstNonEmpty(trade.CreatedAt, trade.LastUpdated),
		TransactionHash: trade.TransactionHash,
	})
	row.TradeCount = len(row.Trades)
	b.report.Summary.CLOBTrades++
	if added {
		b.report.Summary.MatchedNotional += price * size
	}
}

func appendTrade(row *Row, trade TradeSummary) bool {
	tx := strings.TrimSpace(trade.TransactionHash)
	if tx != "" {
		for i, existing := range row.Trades {
			if !strings.EqualFold(existing.TransactionHash, tx) {
				continue
			}
			if existing.Source == SourceData && trade.Source == SourceCLOB {
				row.Trades[i] = trade
			}
			return false
		}
	}
	row.Trades = append(row.Trades, trade)
	return true
}

func (b *reportBuilder) row(market, tokenID string) *Row {
	key := rowKey(market, tokenID)
	if row, ok := b.rows[key]; ok {
		return row
	}
	row := &Row{
		Market:  strings.TrimSpace(market),
		TokenID: strings.TrimSpace(tokenID),
		Status:  StatusUnknown,
	}
	b.rows[key] = row
	b.order = append(b.order, key)
	return row
}

func (b *reportBuilder) build() *Report {
	rows := make([]Row, 0, len(b.order))
	for _, key := range b.order {
		row := b.rows[key]
		row.TradeCount = len(row.Trades)
		row.OpenOrderCount = len(row.OpenOrders)
		rows = append(rows, *row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Status != rows[j].Status {
			return statusRank(rows[i].Status) < statusRank(rows[j].Status)
		}
		return rows[i].Title < rows[j].Title
	})
	b.report.Rows = rows
	return &b.report
}

func (b *reportBuilder) countStatus(status string) {
	switch status {
	case StatusWon:
		b.report.Summary.Won++
	case StatusLost:
		b.report.Summary.Lost++
	case StatusOpen:
		b.report.Summary.Open++
	case StatusClosed:
		b.report.Summary.Closed++
	default:
		b.report.Summary.Unknown++
	}
}

func classifyPosition(position types.Position) string {
	if position.Redeemable {
		if position.CurrentPrice >= 0.999 || position.CurrentValue > position.InitialValue || position.CashPnl > 0 {
			return StatusWon
		}
		if position.CurrentPrice <= 0.001 || position.CurrentValue == 0 || position.CashPnl < 0 {
			return StatusLost
		}
	}
	if position.Size > 0 {
		return StatusOpen
	}
	return StatusUnknown
}

func rowKey(market, tokenID string) string {
	tokenID = strings.TrimSpace(tokenID)
	if tokenID != "" {
		return "token:" + tokenID
	}
	market = strings.TrimSpace(market)
	if market != "" {
		return "market:" + market
	}
	return "unknown"
}

func emptyPosition(position types.Position) bool {
	return position.TokenID == "" && position.ConditionID == "" && position.Size == 0 && position.Title == ""
}

func emptyClosedPosition(position types.ClosedPosition) bool {
	return position.TokenID == "" && position.ConditionID == "" && position.Size == 0 && position.RealizedPnl == 0 && position.Title == ""
}

func emptyDataTrade(trade types.Trade) bool {
	return trade.ID == "" && trade.Market == "" && trade.AssetID == "" && trade.Size == 0
}

func emptyCLOBOrder(order sdkclob.OrderRecord) bool {
	return order.ID == "" && order.Market == "" && order.AssetID == ""
}

func emptyCLOBTrade(trade sdkclob.TradeRecord) bool {
	return trade.ID == "" && trade.Market == "" && trade.AssetID == ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonZero(values ...float64) float64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func parseFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
}

func statusRank(status string) int {
	switch status {
	case StatusWon:
		return 0
	case StatusLost:
		return 1
	case StatusOpen:
		return 2
	case StatusClosed:
		return 3
	default:
		return 4
	}
}
