package gmo

import (
	"log"
)

const (
	StatusDir           = "/v1/status"
	TickerDir           = "/v1/ticker"
	CandlesDir          = "/v1/klines"
	MarginDir           = "/v1/account/margin"
	AssetsDir           = "/v1/account/assets"
	OrdersDir           = "/v1/orders"
	ActiveOrdersDir     = "/v1/activeOrders"
	ExecutionsDir       = "/v1/executions"
	LatestExecutionsDir = "/v1/latestExecutions"
	PositionsDir        = "/v1/openPositions"
	PositionSummaryDir  = "/v1/positionSummary"
	OpenDir             = "/v1/order"
	ChangeOrderDir      = "/v1/changeOrder"
	CancelOrderDir      = "/v1/cancelOrder"
	CancelAllDir        = "/v1/cancelBulkOrder"
	CloseDir            = "/v1/closeOrder"
	CloseAllDir         = "/v1/closeBulkOrder"
)

func pSymbol(param map[string]string, sym string) {
	param["symbol"] = sym
}

func pSymboli(param map[string]interface{}, sym string) {
	param["symbol"] = sym
}

func pCandles(param map[string]string, sym, itv, date string) {
	pSymbol(param, sym)
	param["interval"] = itv
	param["date"] = date
}

func pSidei(param map[string]interface{}, side string) {
	param["side"] = side
}
func pPricei(param map[string]interface{}, price string) {
	if price != "" {
		param["price"] = price
	}
}
func pSizei(param map[string]interface{}, size string) {
	param["size"] = size
}
func pExecTypei(param map[string]interface{}, etype string) {
	param["executionType"] = etype
}

func pLossCutPricei(param map[string]interface{}, price string) {
	if price != "" {
		param["losscutPrice"] = price
	}
}

func pTimeInForcei(param map[string]interface{}, tif string) {
	if tif != "" {
		param["timeInForce"] = tif
	}
}

func pPositionIdi(param map[string]interface{}, posId string) {
	param["positionId"] = posId
}

func pOrderIdi(param map[string]interface{}, orderId string) {
	param["orderId"] = orderId
}

func pSettlei(param map[string]interface{}, settle string) {
	//settle -> "OPEN" "CLOSE"
	if settle != "" {
		param["settleType"] = settle
	}
}

func pOrderId(param map[string]string, orderId string) {
	param["orderId"] = orderId
}

func pExecutionId(param map[string]string, executionId string) {
	param["executionId"] = executionId
}

func pEitherId(param map[string]string, id1, id2 string) {
	if id1 == "" && id2 == "" {
		log.Fatal("id1 or id2 must be passed.")
	}
	if id1 != "" {
		pOrderId(param, id1)
	} else if id2 != "" {
		pExecutionId(param, id2)
	}
}

func pPage(param map[string]string, page, count string) {
	if page != "" {
		param["page"] = page
	}
	if count != "" {
		param["count"] = count
	}
}

func NewStatus(r *ReqHandler) *StatusRes {
	res := &StatusRes{}
	r.Get(StatusDir, nil, res)
	return res
}

func NewTicker(r *ReqHandler, sym string) *TickerRes {
	res := &TickerRes{}
	param := map[string]string{}
	pSymbol(param, sym)
	r.Get(TickerDir, param, res)
	return res
}

func NewCandles(r *ReqHandler, sym, itv, date string) *CandlesRes {
	res := &CandlesRes{}
	param := map[string]string{}
	pCandles(param, sym, itv, date)
	r.Get(CandlesDir, param, res)
	return res
}

func NewMargin(r *ReqHandler) *MarginRes {
	res := &MarginRes{}
	r.GetAuth(MarginDir, nil, res)
	return res
}

func NewAssets(r *ReqHandler) *AssetsRes {
	res := &AssetsRes{}
	r.GetAuth(AssetsDir, nil, res)
	return res
}

func NewOrders(r *ReqHandler, orderId string) *OrdersRes {
	res := &OrdersRes{}
	param := map[string]string{}
	pOrderId(param, orderId)
	r.GetAuth(OrdersDir, param, res)
	return res
}

func NewActiveOrders(r *ReqHandler, sym, page, count string) *OrdersRes {
	res := &OrdersRes{}
	param := map[string]string{}
	pSymbol(param, sym)
	pPage(param, page, count)
	r.GetAuth(ActiveOrdersDir, param, res)
	return res
}

func NewExecutions(r *ReqHandler, orderId, executionId string) *ExecutionsRes {
	res := &ExecutionsRes{}
	param := map[string]string{}
	pEitherId(param, orderId, executionId)
	r.GetAuth(ExecutionsDir, param, res)
	return res
}

func NewLatestExecutions(r *ReqHandler, sym, page, count string) *ExecutionsRes {
	res := &ExecutionsRes{}
	param := map[string]string{}
	pSymbol(param, sym)
	pPage(param, page, count)
	r.GetAuth(LatestExecutionsDir, param, res)
	return res
}

func NewPositions(r *ReqHandler, sym, page, count string) *PositionsRes {
	res := &PositionsRes{}
	param := map[string]string{}
	pSymbol(param, sym)
	pPage(param, page, count)
	r.GetAuth(PositionsDir, param, res)
	return res
}

func NewSummary(r *ReqHandler, sym string) *PositionSummaryRes {
	res := &PositionSummaryRes{}
	param := map[string]string{}
	pSymbol(param, sym)
	r.GetAuth(PositionSummaryDir, param, res)
	return res
}

//**********POST*************

//新規注文
func NewOpenOrder(r *ReqHandler, sym, side, etype,
	price, size, losscut, tif string) *OpenRes {

	res := &OpenRes{}
	param := map[string]interface{}{}
	pSymboli(param, sym)
	pSidei(param, side)
	pExecTypei(param, etype)
	pPricei(param, price)
	pSizei(param, size)
	pLossCutPricei(param, losscut)
	pTimeInForcei(param, tif)
	r.Post(OpenDir, param, res)
	return res
}

//決済
func NewCloseOrder(r *ReqHandler, sym, side, etype,
	price, posId, size, tif string) *CloseRes {

	res := &CloseRes{}
	param := map[string]interface{}{}
	inParam := map[string]interface{}{}
	//main param
	pSymboli(param, sym)
	pSidei(param, side)
	pExecTypei(param, etype)
	pPricei(param, price)
	pTimeInForcei(param, tif)
	//nested param
	pPositionIdi(inParam, posId)
	pSizei(inParam, size)
	//nest param
	param["settlePosition"] = []interface{}{inParam}
	r.Post(CloseDir, param, res)
	return res
}

func NewCloseAll(r *ReqHandler, sym, side, etype, price, size, tif string) *CloseAllRes {
	res := &CloseAllRes{}
	param := map[string]interface{}{}
	pSymboli(param, sym)
	pSidei(param, side)
	pExecTypei(param, etype)
	pSizei(param, size)
	pPricei(param, price)
	pTimeInForcei(param, tif)
	r.Post(CloseAllDir, param, res)
	return res
}

func NewChangeOrder(r *ReqHandler, orderId, price, losscut string) *ChangeOrderRes {
	res := &ChangeOrderRes{}
	param := map[string]interface{}{}
	pOrderIdi(param, orderId)
	pPricei(param, price)
	pLossCutPricei(param, losscut)
	r.Post(ChangeOrderDir, param, res)
	return res
}

func NewCancelOrder(r *ReqHandler, orderId string) *CancelOrderRes {
	res := &CancelOrderRes{}
	param := map[string]interface{}{}
	pOrderIdi(param, orderId)
	r.Post(CancelOrderDir, param, res)
	return res
}

func NewCancelAll(r *ReqHandler, sym, side, settle string) *CancelAllRes {
	res := &CancelAllRes{}
	param := map[string]interface{}{}
	param["symbols"] = []string{sym}
	pSidei(param, side)
	pSettlei(param, settle)
	r.Post(CancelAllDir, param, res)
	return res
}
