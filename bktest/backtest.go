package backtest

import (
	"encoding/json"
	"example/zen/gmo5/fibo"
	"example/zen/gmo5/gmo"
	"example/zen/gmo5/minmax"
	"fmt"
	"os"
)

const (
	SYMBOL = "BTC_JPY"
)

func marketOpen(r *gmo.ReqHandler, side, size string) *gmo.OpenRes {
	res := gmo.NewOpenOrder(r, SYMBOL, side, "MARKET", "", size, "", "")
	return res
}

func marketClose(r *gmo.ReqHandler, side, size, posId string) *gmo.CloseRes {
	res := gmo.NewCloseOrder(r, SYMBOL, side, "MARKET", "", posId, size, "")
	return res
}

//LIMIT or STOP open-order.
//etype = "LIMIT" or "STOP"
func limitOpen(r *gmo.ReqHandler, side, etype, size, price string) *gmo.OpenRes {
	res := gmo.NewOpenOrder(r, SYMBOL, side, etype, price, size, "", "")
	return res
}

//LIMIT or STOP close-order.
//etype = "LIMIT" or "STOP"
func limitClose(r *gmo.ReqHandler, side, etype, size, price, posId string) *gmo.CloseRes {
	res := gmo.NewCloseOrder(r, SYMBOL, side, etype, price, posId, size, "")
	return res
}

func newTicker(r *gmo.ReqHandler) *gmo.TickerData {
	if res := gmo.NewTicker(r, SYMBOL); res == nil {
		return nil
	} else {
		return res.Extract(SYMBOL)
	}
}

//test types

type Chart struct {
	X      []int
	Y      []float64
	Side   []string //"BUY"" or "S"ELL"
	Action []string //"OPEN" or "CLOSE"
}

type Balance struct {
	X []int
	Y []float64
}

type Position struct {
	size  float64
	price float64
	side  string
	time  int
}

type Summary struct {
	Position
	chart  Chart
	pl     float64 //total profit
	profR  float64 // profit ratio
	lossR  float64 //loss ratio MUST BE NEGATIVE
	spread float64
	cnt    int // count of trades
}

//methods

func (b *Balance) add(ot int, v float64) {
	b.X = append(b.X, ot)
	b.Y = append(b.Y, v)
}

func (b *Balance) write(fpath string) {
	if b, err := json.MarshalIndent(b, "", " "); err == nil {
		os.WriteFile(fpath, b, 0777)
	}
}

func (ch *Chart) add(ot int, v float64, side, act string) {
	ch.X = append(ch.X, ot)
	ch.Y = append(ch.Y, v)
	ch.Side = append(ch.Side, side)
	ch.Action = append(ch.Action, act)
}
func (ch *Chart) write(fpath string) {
	if b, err := json.MarshalIndent(ch, "", " "); err == nil {
		os.WriteFile(fpath, b, 0777)
	}
}

func (p *Position) has() bool {
	return p.size != 0.0
}

func (p *Position) check(v float64) float64 {
	if p.side == "BUY" {
		return (v - p.price) * p.size
	} else {
		return (p.price - v) * p.size
	}
}

func (s *Summary) isProfFilled(v float64) bool {
	if !s.Position.has() {
		return false
	}
	if s.side == "BUY" {
		return (v-s.price)/s.price >= s.profR
	} else {
		return (s.price-v)/s.price >= s.profR
	}
}
func (s *Summary) isLossFilled(v float64) bool {
	if !s.Position.has() {
		return false
	}
	if s.side == "BUY" {
		return (v-s.price)/s.price <= s.lossR
	} else {
		return (s.price-v)/s.price <= s.lossR
	}
}

func (s *Summary) open(price, size float64, otime int, side string) {
	s.price = price
	s.size = size
	s.side = side
	s.time = otime
	s.chart.add(otime, price, side, "OPEN")
}
func (s *Summary) close(price float64, otime int) float64 {
	var pl float64
	if s.side == "BUY" {
		pl = (price - s.price - s.spread) * s.size
		s.pl += pl
	} else {
		pl = (s.price - price - s.spread) * s.size
		s.pl += pl
	}
	//add to chart
	var side string
	if s.side == "BUY" {
		side = "SELL"
	} else {
		side = "BUY"
	}

	//print infos
	//fmt.Printf("CLOSE-%v! open:%v price:%.f close:%v price:%.f prof:%.f\n",s.side,s.time, s.price, otime, price,pl)
	s.chart.add(otime, price, side, "CLOSE")
	s.cnt++

	//reset
	s.price = 0.0
	s.size = 0.0

	s.side = ""
	s.time = 0
	return pl
}

func Backtest() {
	//read test data
	tdata := &gmo.CandlesData{}
	if b, err := os.ReadFile("./testdata4hr.json"); err == nil {
		if err := json.Unmarshal(b, tdata); err != nil {
			fmt.Println(err)
		}
	}
	tsize := 0.1
	//lcount := 42 // 7days (4hr)
	mcount := 80 // 5days (4hr)

	pos := &Summary{profR: 0.05, lossR: -0.05, spread: 1200.00}
	bal := &Balance{}
	offset := 0
	for i, v := range tdata.Close[offset+mcount+1:] {
		otime := tdata.OpenTime[offset+mcount+i+1]
		ed := offset + i + mcount
		st := ed - mcount
		inf := minmax.NewInf(tdata.High[st:ed], tdata.Low[st:ed]).AddWrap(v)
		var dec string

		dec = "" //open判定変数

		//最大値更新中
		if v >= inf.Maxv {
			if !pos.has() {
				//ポジションないときはBUY
				//pos.open(v, tsize, otime, "BUY")
				dec = "BUY"
			} else if pos.side == "SELL" {
				//売りポジ持っているときは決済して再購入
				pl := pos.close(v, otime)
				_ = pl
				//fmt.Printf("trenChange:%.f\n", pl)
				//pos.open(v, tsize, otime, "BUY")
				dec = "BUY"
			}
		} else if v <= inf.Minv {
			if !pos.has() {
				//pos.open(v, tsize, otime, "SELL")
				dec = "SELL"
			} else if pos.side == "BUY" {
				pl := pos.close(v, otime)
				_ = pl
				//fmt.Printf("trendChange:%.f\n", pl)
				//pos.open(v, tsize, otime, "SELL")
				dec = "SELL"
			}
		}

		//損失満たす
		if pos.isLossFilled(v) {
			pl := pos.close(v, otime)
			_ = pl
			//fmt.Printf("losscut:%.f\n", pl)
		}

		//利益満たす
		if pos.isProfFilled(v) {
			pl := pos.close(v, otime)
			_ = pl
			//fmt.Printf("prof:%.f\n", pl)
		}

		//open判定が設定されていない場合、
		//fibolevelが5以上もしくは１以下の場合設定
		if dec == "" {
			lvl := fibo.Level(inf.Scaled)
			if lvl >= 5 {
				//fibo 76.4%以上で順張り
				if inf.Which == "B" {
					dec = "BUY"
				} else if inf.Which == "T" {
					dec = "SELL"
				}
			}
			if lvl <= 1 {
				if inf.Which == "B" {
					dec = "BUY"
				} else if inf.Which == "T" {
					dec = "SELL"
				}
			}
		}

		//open判定かつポジション無しなれopen。
		if dec != "" && !pos.has() {
			pos.open(v, tsize, otime, dec)
		}

		//balance更新
		bal.add(otime, pos.pl)
	}
	fmt.Printf("prof:%.f trades:%v\n", pos.pl, pos.cnt)
	pos.chart.write("./chartdata/pos.json")
	bal.write("./chartdata/bal.json")
}
