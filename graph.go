package main

/*
AddTradeHistory())
AddBalance()
AddCandleData()
*/

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zenryokukun/surfergopher/gmo"
)

const (
	CDATA_FPATH = "./data/candle.json"  //グラフ用　ろうそく足データ
	BDATA_FPATH = "./data/balance.json" //グラフ用　残高推移
	TRADE_FPATH = "./data/trade.json"   //グラフ用　取引履歴
)

type (
	XY struct {
		X []int     //gmo api responseのOpenTimeを想定
		Y []float64 //価格
	}

	History struct {
		XY
		Side   []string //"BUY"" or "S"ELL"
		Action []string //"OPEN" or "CLOSE"
	}
)

func NewHistory(x int, y float64, side, action string) *History {
	return &History{
		XY: XY{
			X: []int{x}, Y: []float64{y},
		},
		Side: []string{side}, Action: []string{action},
	}
}

//総利益をファイルに出力
func AddBalance(x int, y float64) {
	xy := &XY{}
	b, err := os.ReadFile(BDATA_FPATH)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(b, xy)
	if err != nil {
		fmt.Println(err)
		//return
	}
	xy.addXY(x, y)
	xy.write()
}

//取引情報をファイルに出力
func AddTradeHistory(nh *History) {
	histo := &History{}
	b, err := os.ReadFile(TRADE_FPATH)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(b, histo)
	if err != nil {
		fmt.Println(err)
		//return
	}
	histo.merge(nh)
	histo.write()
}

//ロウソク足をファイルに出力
func AddCandleData(cd *gmo.CandlesData) {
	b, err := json.MarshalIndent(cd, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create(CDATA_FPATH)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write(b)
}

func (xy *XY) addXY(x int, y float64) {
	xy.X = append(xy.X, x)
	xy.Y = append(xy.Y, y)
}

func (xy *XY) merge(nxy *XY) {
	xy.X = append(xy.X, nxy.X...)
	xy.Y = append(xy.Y, nxy.Y...)
}

func (xy *XY) slice(start int) {
	xy.X = xy.X[start:]
	xy.Y = xy.Y[start:]
}

func (xy *XY) write() {
	f, err := os.Create(BDATA_FPATH)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	b, err := json.MarshalIndent(xy, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write(b)
}

func (h *History) addHistory(x int, y float64, side, act string) {
	h.addXY(x, y)
	h.Side = append(h.Side, side)
	h.Action = append(h.Action, act)
}

func (h *History) merge(nh *History) {
	h.X = append(h.X, nh.X...)
	h.Y = append(h.Y, nh.Y...)
	h.Side = append(h.Side, nh.Side...)
	h.Action = append(h.Action, nh.Action...)
}

func (h *History) slice(start int) {
	h.XY.slice(start)
	h.Side = h.Side[start:]
	h.Action = h.Action[start:]
}

func (h *History) write() {
	f, err := os.Create(TRADE_FPATH)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	b, err := json.MarshalIndent(h, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write(b)
}

func writeData(v interface{}, fpath string) {
	f, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write(b)
}
