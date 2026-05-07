package rfq

import (
	"fmt"
)

// HTTPClientInterface HTTP客户端接口
type HTTPClientInterface interface {
	Get(path string, headers map[string]string) (interface{}, error)
	Post(path string, headers map[string]string, body interface{}) (interface{}, error)
	Delete(path string, headers map[string]string, body interface{}) (interface{}, error)
}

// SignedOrderData 签名订单数据（用于避免循环导入）
type SignedOrderData struct {
	Salt          int64  `json:"salt"`
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenID       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          string `json:"side"`
	SignatureType int    `json:"signatureType"`
	Signature     string `json:"signature"`
}

// OrderCreationArgs 订单创建参数
type OrderCreationArgs struct {
	TokenID    string
	Price      float64
	Size       float64
	Side       string
	Expiration int
}

// ClobClientInterface CLOB客户端接口（避免循环导入）
type ClobClientInterface interface {
	AssertLevel2Auth() error
	GetHTTPClient() HTTPClientInterface
	GetHost() string
	CreateLevel2HeadersInternal(method, path string, body interface{}) (map[string]string, error)
	GetAPICreds() (apiKey string)
	CreateOrderForRFQ(args *OrderCreationArgs) (*SignedOrderData, error)
}

// RfqClient RFQ客户端
type RfqClient struct {
	parent ClobClientInterface
}

// NewRfqClient 创建新的RFQ客户端
func NewRfqClient(parent ClobClientInterface) *RfqClient {
	return &RfqClient{
		parent: parent,
	}
}

// ensureL2Auth 确保L2认证
func (r *RfqClient) ensureL2Auth() error {
	return r.parent.AssertLevel2Auth()
}

// getL2Headers 获取L2认证头
func (r *RfqClient) getL2Headers(method, endpoint string, body interface{}) (map[string]string, error) {
	return r.parent.CreateLevel2HeadersInternal(method, endpoint, body)
}

// CreateRfqRequest 创建RFQ请求
func (r *RfqClient) CreateRfqRequest(request *RfqUserRequest) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	headers, err := r.getL2Headers("POST", "/rfq/request", request)
	if err != nil {
		return nil, err
	}

	httpClient := r.parent.GetHTTPClient()
	return httpClient.Post("/rfq/request", headers, request)
}

// CancelRfqRequest 取消RFQ请求
func (r *RfqClient) CancelRfqRequest(params *CancelRfqRequestParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	headers, err := r.getL2Headers("DELETE", "/rfq/request", params)
	if err != nil {
		return nil, err
	}

	return r.parent.GetHTTPClient().Delete("/rfq/request", headers, params)
}

// GetRfqRequests 获取RFQ请求列表
func (r *RfqClient) GetRfqRequests(params *GetRfqRequestsParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	// 构建查询参数
	path := "/rfq/data/requests"
	if params != nil {
		path += "?"
		if params.TokenID != "" {
			path += fmt.Sprintf("token_id=%s&", params.TokenID)
		}
		if params.Side != "" {
			path += fmt.Sprintf("side=%s&", params.Side)
		}
		if params.Status != "" {
			path += fmt.Sprintf("status=%s&", params.Status)
		}
		// 移除末尾的&
		if len(path) > 0 && path[len(path)-1] == '&' {
			path = path[:len(path)-1]
		}
	}

	headers, err := r.getL2Headers("GET", "/rfq/data/requests", nil)
	if err != nil {
		return nil, err
	}

	httpClient := r.parent.GetHTTPClient()
	return httpClient.Get(path, headers)
}

// CreateRfqQuote 创建RFQ报价
func (r *RfqClient) CreateRfqQuote(quote *RfqUserQuote) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	headers, err := r.getL2Headers("POST", "/rfq/quote", quote)
	if err != nil {
		return nil, err
	}

	return r.parent.GetHTTPClient().Post("/rfq/quote", headers, quote)
}

// CancelRfqQuote 取消RFQ报价
func (r *RfqClient) CancelRfqQuote(params *CancelRfqQuoteParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	headers, err := r.getL2Headers("DELETE", "/rfq/quote", params)
	if err != nil {
		return nil, err
	}

	return r.parent.GetHTTPClient().Delete("/rfq/quote", headers, params)
}

// GetRfqQuotes 获取RFQ报价列表（旧接口，建议使用 GetRfqRequesterQuotes 或 GetRfqQuoterQuotes）
// Deprecated: 使用 GetRfqRequesterQuotes 或 GetRfqQuoterQuotes
func (r *RfqClient) GetRfqQuotes(params *GetRfqQuotesParams) (interface{}, error) {
	// 默认使用 requester 视角
	return r.GetRfqRequesterQuotes(params)
}

// GetRfqRequesterQuotes 获取针对自己请求的报价列表（请求方视角）
// 返回别人对你的 RFQ 请求做出的报价
func (r *RfqClient) GetRfqRequesterQuotes(params *GetRfqQuotesParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	// 构建查询参数
	path := "/rfq/data/requester/quotes"
	if params != nil {
		path += "?"
		if params.RequestID != "" {
			path += fmt.Sprintf("request_id=%s&", params.RequestID)
		}
		if params.TokenID != "" {
			path += fmt.Sprintf("token_id=%s&", params.TokenID)
		}
		if params.Side != "" {
			path += fmt.Sprintf("side=%s&", params.Side)
		}
		if params.Status != "" {
			path += fmt.Sprintf("status=%s&", params.Status)
		}
		if len(params.QuoteIDs) > 0 {
			for _, qid := range params.QuoteIDs {
				path += fmt.Sprintf("quoteIds=%s&", qid)
			}
		}
		// 移除末尾的&
		if len(path) > 0 && path[len(path)-1] == '&' {
			path = path[:len(path)-1]
		}
	}

	headers, err := r.getL2Headers("GET", "/rfq/data/requester/quotes", nil)
	if err != nil {
		return nil, err
	}

	httpClient := r.parent.GetHTTPClient()
	return httpClient.Get(path, headers)
}

// GetRfqQuoterQuotes 获取自己创建的报价列表（报价方视角）
// 返回你对别人 RFQ 请求做出的报价
func (r *RfqClient) GetRfqQuoterQuotes(params *GetRfqQuotesParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	// 构建查询参数
	path := "/rfq/data/quoter/quotes"
	if params != nil {
		path += "?"
		if params.RequestID != "" {
			path += fmt.Sprintf("request_id=%s&", params.RequestID)
		}
		if params.TokenID != "" {
			path += fmt.Sprintf("token_id=%s&", params.TokenID)
		}
		if params.Side != "" {
			path += fmt.Sprintf("side=%s&", params.Side)
		}
		if params.Status != "" {
			path += fmt.Sprintf("status=%s&", params.Status)
		}
		if len(params.QuoteIDs) > 0 {
			for _, qid := range params.QuoteIDs {
				path += fmt.Sprintf("quoteIds=%s&", qid)
			}
		}
		// 移除末尾的&
		if len(path) > 0 && path[len(path)-1] == '&' {
			path = path[:len(path)-1]
		}
	}

	headers, err := r.getL2Headers("GET", "/rfq/data/quoter/quotes", nil)
	if err != nil {
		return nil, err
	}

	httpClient := r.parent.GetHTTPClient()
	return httpClient.Get(path, headers)
}

// GetRfqBestQuote 获取最佳RFQ报价
func (r *RfqClient) GetRfqBestQuote(params *GetRfqBestQuoteParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/rfq/data/best-quote?token_id=%s&side=%s&size=%.6f",
		params.TokenID, params.Side, params.Size)

	headers, err := r.getL2Headers("GET", "/rfq/data/best-quote", nil)
	if err != nil {
		return nil, err
	}

	httpClient := r.parent.GetHTTPClient()
	return httpClient.Get(path, headers)
}

// AcceptQuote 接受报价（请求方）
// 此方法会获取报价详情，创建签名订单，然后提交接受请求
func (r *RfqClient) AcceptQuote(params *AcceptQuoteParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	// 步骤1: 获取报价详情（使用 requester 视角）
	quotesResp, err := r.GetRfqRequesterQuotes(&GetRfqQuotesParams{
		QuoteIDs: []string{params.QuoteID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get RFQ quotes: %w", err)
	}

	// 解析响应
	respMap, ok := quotesResp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	dataList, ok := respMap["data"].([]interface{})
	if !ok || len(dataList) == 0 {
		return nil, fmt.Errorf("RFQ quote not found")
	}

	// 找到对应的报价
	var rfqQuote map[string]interface{}
	for _, item := range dataList {
		quote, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if quote["quoteId"] == params.QuoteID {
			rfqQuote = quote
			break
		}
	}
	if rfqQuote == nil {
		return nil, fmt.Errorf("RFQ quote with ID %s not found", params.QuoteID)
	}

	// 步骤2: 构建订单创建参数
	orderCreationPayload, err := r.getRequestOrderCreationPayload(rfqQuote)
	if err != nil {
		return nil, fmt.Errorf("failed to get order creation payload: %w", err)
	}

	priceStr, _ := rfqQuote["price"].(string)
	price := 0.0
	fmt.Sscanf(priceStr, "%f", &price)

	orderArgs := &OrderCreationArgs{
		TokenID:    orderCreationPayload.Token,
		Price:      price,
		Size:       orderCreationPayload.Size,
		Side:       orderCreationPayload.Side,
		Expiration: params.Expiration,
	}

	// 步骤3: 创建签名订单
	order, err := r.parent.CreateOrderForRFQ(orderArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 步骤4: 构建接受请求的payload
	acceptPayload := map[string]interface{}{
		"requestId":     params.RequestID,
		"quoteId":       params.QuoteID,
		"owner":         r.parent.GetAPICreds(),
		"salt":          order.Salt,
		"maker":         order.Maker,
		"signer":        order.Signer,
		"taker":         order.Taker,
		"tokenId":       order.TokenID,
		"makerAmount":   order.MakerAmount,
		"takerAmount":   order.TakerAmount,
		"expiration":    order.Expiration,
		"nonce":         order.Nonce,
		"feeRateBps":    order.FeeRateBps,
		"side":          orderCreationPayload.Side,
		"signatureType": order.SignatureType,
		"signature":     order.Signature,
	}

	headers, err := r.getL2Headers("POST", "/rfq/request/accept", acceptPayload)
	if err != nil {
		return nil, err
	}

	return r.parent.GetHTTPClient().Post("/rfq/request/accept", headers, acceptPayload)
}

// ApproveOrder 批准订单（报价方）
// 此方法会获取报价详情，创建签名订单，然后提交批准请求
func (r *RfqClient) ApproveOrder(params *ApproveOrderParams) (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	// 步骤1: 获取报价详情（使用 quoter 视角）
	quotesResp, err := r.GetRfqQuoterQuotes(&GetRfqQuotesParams{
		QuoteIDs: []string{params.QuoteID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get RFQ quotes: %w", err)
	}

	// 解析响应
	respMap, ok := quotesResp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	dataList, ok := respMap["data"].([]interface{})
	if !ok || len(dataList) == 0 {
		return nil, fmt.Errorf("RFQ quote not found")
	}

	// 找到对应的报价
	var rfqQuote map[string]interface{}
	for _, item := range dataList {
		quote, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if quote["quoteId"] == params.QuoteID {
			rfqQuote = quote
			break
		}
	}
	if rfqQuote == nil {
		return nil, fmt.Errorf("RFQ quote with ID %s not found", params.QuoteID)
	}

	// 步骤2: 根据报价详情创建订单
	// 报价方使用自己报价的 side
	side, _ := rfqQuote["side"].(string)
	if side == "" {
		side = "BUY"
	}

	// 根据 side 确定 size
	var size float64
	if side == "BUY" {
		sizeStr, _ := rfqQuote["sizeIn"].(string)
		fmt.Sscanf(sizeStr, "%f", &size)
	} else {
		sizeStr, _ := rfqQuote["sizeOut"].(string)
		fmt.Sscanf(sizeStr, "%f", &size)
	}

	tokenID, _ := rfqQuote["token"].(string)
	priceStr, _ := rfqQuote["price"].(string)
	price := 0.0
	fmt.Sscanf(priceStr, "%f", &price)

	orderArgs := &OrderCreationArgs{
		TokenID:    tokenID,
		Price:      price,
		Size:       size,
		Side:       side,
		Expiration: params.Expiration,
	}

	// 步骤3: 创建签名订单
	order, err := r.parent.CreateOrderForRFQ(orderArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 步骤4: 构建批准请求的payload
	approvePayload := map[string]interface{}{
		"requestId":     params.RequestID,
		"quoteId":       params.QuoteID,
		"owner":         r.parent.GetAPICreds(),
		"salt":          order.Salt,
		"maker":         order.Maker,
		"signer":        order.Signer,
		"taker":         order.Taker,
		"tokenId":       order.TokenID,
		"makerAmount":   order.MakerAmount,
		"takerAmount":   order.TakerAmount,
		"expiration":    order.Expiration,
		"nonce":         order.Nonce,
		"feeRateBps":    order.FeeRateBps,
		"side":          side,
		"signatureType": order.SignatureType,
		"signature":     order.Signature,
	}

	headers, err := r.getL2Headers("POST", "/rfq/quote/approve", approvePayload)
	if err != nil {
		return nil, err
	}

	return r.parent.GetHTTPClient().Post("/rfq/quote/approve", headers, approvePayload)
}

// GetRfqConfig 获取RFQ配置
func (r *RfqClient) GetRfqConfig() (interface{}, error) {
	if err := r.ensureL2Auth(); err != nil {
		return nil, err
	}

	headers, err := r.getL2Headers("GET", "/rfq/config", nil)
	if err != nil {
		return nil, err
	}

	return r.parent.GetHTTPClient().Get("/rfq/config", headers)
}

// OrderCreationResult 订单创建结果
type OrderCreationResult struct {
	Token string
	Side  string
	Size  float64
}

// getRequestOrderCreationPayload 根据报价详情构建订单创建参数
// 与 Python 的 _get_request_order_creation_payload 对应
func (r *RfqClient) getRequestOrderCreationPayload(quote map[string]interface{}) (*OrderCreationResult, error) {
	rawMatchType, _ := quote["matchType"].(string)
	matchType := MatchType(rawMatchType)
	if matchType == "" {
		matchType = MatchTypeComplementary
	}

	side, _ := quote["side"].(string)
	if side == "" {
		side = "BUY"
	}

	switch matchType {
	case MatchTypeComplementary:
		// 对于 BUY <> SELL 和 SELL <> BUY
		// 订单的 side 与报价的 side 相反
		token, _ := quote["token"].(string)
		if token == "" {
			return nil, fmt.Errorf("missing token for COMPLEMENTARY match")
		}

		// 反转 side
		if side == "BUY" {
			side = "SELL"
		} else {
			side = "BUY"
		}

		var sizeStr string
		if side == "BUY" {
			sizeStr, _ = quote["sizeOut"].(string)
		} else {
			sizeStr, _ = quote["sizeIn"].(string)
		}
		if sizeStr == "" {
			return nil, fmt.Errorf("missing sizeIn/sizeOut for COMPLEMENTARY match")
		}

		size := 0.0
		fmt.Sscanf(sizeStr, "%f", &size)

		return &OrderCreationResult{
			Token: token,
			Side:  side,
			Size:  size,
		}, nil

	case MatchTypeMint, MatchTypeMerge:
		// BUY <> BUY, SELL <> SELL
		// 订单的 side 与报价的 side 相同
		token, _ := quote["complement"].(string)
		if token == "" {
			return nil, fmt.Errorf("missing complement token for MINT/MERGE match")
		}

		var sizeStr string
		if side == "BUY" {
			sizeStr, _ = quote["sizeIn"].(string)
		} else {
			sizeStr, _ = quote["sizeOut"].(string)
		}
		if sizeStr == "" {
			return nil, fmt.Errorf("missing sizeIn/sizeOut for MINT/MERGE match")
		}

		size := 0.0
		fmt.Sscanf(sizeStr, "%f", &size)

		return &OrderCreationResult{
			Token: token,
			Side:  side,
			Size:  size,
		}, nil

	default:
		return nil, fmt.Errorf("invalid match type: %s", rawMatchType)
	}
}
