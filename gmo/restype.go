package gmo

import (
	"fmt"
	"strconv"
)

//variable
const ok = 0

//interface not in use..?
type extractor interface {
	Extract()
}

//response types
type Base struct {
	Status       int                 `json:"status"`
	Msg          []map[string]string `json:"messages"`
	ResponseTime string              `json:"responsetime"`
}

type CancelOrderRes Base

type ChangeOrderRes Base

type StatusRes struct {
	Base
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
}

type Ticker struct {
	Ask       int     `json:"ask,string"`
	Bid       int     `json:"bid,string"`
	High      int     `json:"high,string"`
	Last      int     `json:"last,string"`
	Low       int     `json:"low,string"`
	Symbol    string  `json:"symbol"`
	Timestamp string  `json:"timestamp"`
	Volume    float64 `json:"volume,string"`
}
type TickerRes struct {
	Base
	Data []Ticker `json:"data"`
}

type Candle struct {
	OpenTime string `json:"openTime"`
	Open     string `json:"open"`
	High     string `json:"high"`
	Low      string `json:"low"`
	Close    string `json:"close"`
	Volume   string `json:"volume"`
}
type CandlesRes struct {
	Base
	Data []Candle `json:"data"`
}

type MarginRes struct {
	Base
	Data struct {
		AcutalProfitLoss string `json:"actualProfitLoss"`
		AvailableAmount  string `json:"availableAmount"`
		Margin           string `json:"margin"`
		MarginCallStatus string `json:"marginCallStatus"`
		MarginRatio      string `json:"marginRatio"`
		ProfitLoss       string `json:"profitLoss"`
	} `json:"data"`
}

type AssetsRes struct {
	Base
	Data []struct {
		Amount         string `json:"amount"`
		Available      string `json:"available"`
		ConversionRate string `json:"conversionRate"`
		Symbol         string `json:"symbol"`
	} `json:"data"`
}

type pagination struct {
	Page  int `json:"currentPage"`
	Count int `json:"count"`
}

type orderList struct {
	RootOrderId   uint64 `json:"rootOrderId"`
	OrderId       uint64 `json:"orderId"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	ExecutionType string `json:"executionType"`
	SettleType    string `json:"settleType"`
	Size          string `json:"size"`
	ExecutedSize  string `json:"executedSize"`
	Price         string `json:"price"`
	Status        string `json:"status"`
	TimeInForce   string `json:"timeInForce"`
	TimeStamp     string `json:"timestamp"`
}

type OrdersRes struct {
	//orders,activeOrders兼用
	Base
	Data struct {
		Pagination pagination  `json:"pagination"`
		List       []orderList `json:"list"`
	} `json:"data"`
}

type ExecutionsRes struct {
	//executions,latestExecutions 兼用
	Base
	Data struct {
		Pagination pagination `json:"pagination"`
		List       []struct {
			ExecutionId uint32 `json:"executionId"`
			OrderId     uint32 `json:"orderId"`
			PositionId  uint32 `json:"positionId"`
			Symbol      string `json:"symbol"`
			Side        string `json:"side"`
			SettleType  string `json:"settleType"`
			Size        string `json:"size"`
			Price       string `json:"price"`
			LossGain    string `json:"lossGain"`
			Fee         string `json:"fee"`
			Timestamp   string `json:"timestamp"`
		} `json:"list"`
	} `json:"data"`
}

type Positions struct {
	PositionId   uint32 `json:"positionId"`
	Symbol       string `json:"symbol"`
	Side         string `json:"side"`
	Size         string `json:"size"`
	OrderSize    string `json:"orderSize"`
	Price        string `json:"price"`
	LossGain     string `json:"lossGain"`
	Leverage     string `json:"leverage"`
	LossCutPrice string `json:"losscutPrice"`
	Timestamp    string `json:"timestamp"`
}
type PositionsRes struct {
	Base
	Data struct {
		Pagination pagination  `json:"pagination"`
		List       []Positions `json:"list"`
	} `json:"data"`
}

type PositionSummaryRes struct {
	Base
	Data struct {
		List []struct {
			AveragePositionRate string `json:"averagePositionRate"`
			LossGain            string `json:"positionLossGain"`
			Side                string `json:"side"`
			OrderQuantity       string `json:"sumOrderQuantity"`
			PositionQuantity    string `json:"sumPositionQuantity"`
			Symbol              string `json:"symbol"`
		} `json:"list"`
	} `json:"data"`
}

type OpenRes struct {
	Base
	Data string `json:"data"`
}

type CloseRes OpenRes
type CloseAllRes OpenRes

type CancelAllRes struct {
	Base
	Data []int `json:"data"`
}

//response.Data converted types
type CandlesData struct {
	Open     []float64
	Close    []float64
	High     []float64
	Low      []float64
	OpenTime []int
}

type TickerData struct {
	Ask       float64
	Bid       float64
	Last      float64
	High      float64
	Low       float64
	Symbol    string
	Timestamp string
}

//********************************************************************************
//methods
//********************************************************************************
func (b *Base) ErrorLog() string {
	if b.Msg == nil {
		return fmt.Sprintf("API STATUS:%v TIME:%v MSG:but no error messages....",
			b.Status, b.ResponseTime)
	}
	msg := b.Msg[0]
	code := msg["message_code"]
	text := msg["message_string"]
	return fmt.Sprintf("API STATUS:%v TIME:%v CODE:%v MSG:%v",
		b.Status, b.ResponseTime, code, text)
}

func (c *CandlesRes) Extract() *CandlesData {
	if c.Status != ok {
		return nil
	}

	var open, close, high, low []float64
	var time []int

	for _, cl := range c.Data {
		if of, err := strconv.ParseFloat(cl.Open, 64); err == nil {
			open = append(open, of)
		}
		if cf, err := strconv.ParseFloat(cl.Close, 64); err == nil {
			close = append(close, cf)
		}
		if hf, err := strconv.ParseFloat(cl.High, 64); err == nil {
			high = append(high, hf)
		}
		if lf, err := strconv.ParseFloat(cl.Low, 64); err == nil {
			low = append(low, lf)
		}
		if ot, err := strconv.Atoi(cl.OpenTime); err == nil {
			time = append(time, ot/1000) //レスポンスはミリ秒で返ってくるため秒に変換
		} else {
			fmt.Println(err)
		}
	}
	return &CandlesData{
		High: high, Low: low, Open: open, Close: close, OpenTime: time,
	}
}

func (c *CandlesData) AddBefore(bef *CandlesData) {
	c.Close = append(bef.Close, c.Close...)
	c.High = append(bef.High, c.High...)
	c.Low = append(bef.Low, c.Low...)
	c.Open = append(bef.Open, c.Open...)
	c.OpenTime = append(bef.OpenTime, c.OpenTime...)
}

func (c *CandlesData) AddAfter(aft *CandlesData) {
	c.Close = append(c.Close, aft.Close...)
	c.High = append(c.High, aft.High...)
	c.Low = append(c.Low, aft.Low...)
	c.Open = append(c.Open, aft.Open...)
	c.OpenTime = append(c.OpenTime, aft.OpenTime...)
}

func (c *CandlesData) Slice(st, end int) *CandlesData {
	c.Close = c.Close[st:end]
	c.High = c.High[st:end]
	c.Low = c.Low[st:end]
	c.Open = c.Open[st:end]
	c.OpenTime = c.OpenTime[st:end]
	return c
}

//Ticker型をTickerDataに変換
//Tickerは価格をintで持つためfloat64に変換
func floatTicker(t *Ticker) *TickerData {
	ask := float64(t.Ask)
	bid := float64(t.Bid)
	high := float64(t.High)
	low := float64(t.Low)
	last := float64(t.Last)
	return &TickerData{
		Ask: ask, Bid: bid, High: high, Low: low, Last: last,
		Timestamp: t.Timestamp, Symbol: t.Symbol,
	}
}

func (t *TickerRes) Extract(sym string) *TickerData {
	if t.Status != ok {
		return nil
	}
	for _, v := range t.Data {
		if v.Symbol == sym {
			return floatTicker(&v)
		}
	}
	return nil
}

func (t *TickerData) Spread() float64 {
	return t.Ask - t.Bid
}

//posIdに対応するorderIdを返す
func (e *ExecutionsRes) FindByPosId(posId string) string {
	if e.Status != ok {
		return ""
	}
	for _, v := range e.Data.List {
		if fmt.Sprint(v.PositionId) == posId {
			return fmt.Sprint(v.OrderId)
		}
	}
	return ""
}

//orderIdに対応するpositionIdを返す
func (e *ExecutionsRes) FindByOrderId(orderId string) string {
	if e.Status != ok {
		return ""
	}
	for _, v := range e.Data.List {
		if fmt.Sprint(v.OrderId) == orderId {
			return fmt.Sprint(v.PositionId)
		}
	}
	return ""
}
