package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/zenryokukun/gotweet"
	"github.com/zenryokukun/surfergopher/bktest"
	"github.com/zenryokukun/surfergopher/gmo"
	"github.com/zenryokukun/surfergopher/minmax"

	_ "github.com/mattn/go-sqlite3"
)

const (
	TEST_MODE      = false // 本番かテストモードか
	GLOBAL_FPATH   = "./globals.json"
	CLOSEDID_FPATH = "./data/closedposlist.txt" // closeしたorderIdのリスト
	CDATA_FPATH    = "./data/candle.json"       // ロウソク足を出力するファイル
	BOTNAME        = "Surfer Gopher"            // botの名前
	VER            = "@v2.0.0"                  // botのversion
	PYSCRIPT       = "./py/chart.py"            // pythonスクリプト
	IMG_PATH       = "./py/tweet.png"           // tweetする画像のパス
	DB_PATH        = "./data.db"                // sqlite3 dbファイルのパス
)

// 環境によって分けるためグローバル変数にした
var (
	SYMBOL         string
	TRADE_INTERVAL string
	TSIZE          string
	PROF_RATIO     float64
	LOSS_RATIO     float64
	SPREAD_THRESH  float64
	TSIZE_F        float32
)

// グローバル変数をファイルから設定
func setGlobalVars() {
	type gl struct {
		Symbol        string  `json:"symbol"`
		TradeInterval string  `json:"tradeInterval"`
		Tsize         string  `json:"tsize"`
		ProfRatio     float64 `json:"profRatio"`
		LossRatio     float64 `json:"lossRatio"`
		SpreadThresh  float64 `json:"spreadThresh"`
	}

	gdata := &gl{}

	if b, err := os.ReadFile(GLOBAL_FPATH); err == nil {
		json.Unmarshal(b, gdata)
	}
	SYMBOL = gdata.Symbol
	TRADE_INTERVAL = gdata.TradeInterval
	PROF_RATIO = gdata.ProfRatio
	LOSS_RATIO = gdata.LossRatio
	TSIZE = gdata.Tsize
	SPREAD_THRESH = gdata.SpreadThresh
	amt, err := strconv.ParseFloat(TSIZE, 32)
	if err != nil {
		logger(fmt.Sprintf("could not parse TSIZE to float:%v\n", amt))
	}
	TSIZE_F = float32(amt)
}

// Returns orderId
func marketOpen(r *gmo.ReqHandler, side, size string) string {
	if waitForSpread(r) {
		//res := _marketOpen(r, side, size)
		res := gmo.NewOpenOrder(r, SYMBOL, side, "MARKET", "", size, "", "")
		if res != nil {
			return res.Data
		} else {
			logger("marketOpen failed.Response was nil.")
		}
	}
	return ""
}

// Returns closeordeId
func marketCloseSide(r *gmo.ReqHandler, p *gmo.Summary) string {
	tradeSide := oppositeSide(p.Side)
	size := fmt.Sprint(p.PositionQuantity)
	if waitForSpread(r) {
		res := gmo.NewCloseAll(r, SYMBOL, tradeSide, "MARKET", "", size, "")
		if res != nil {
			return res.Data
		} else {
			logger("marketCloseSide failed.Response was nil.")
		}
	}
	return ""
}

// Returns slice of closeOrderIds
func marketCloseBoth(r *gmo.ReqHandler, pos []gmo.Summary) []string {
	ids := []string{}
	for _, p := range pos {
		tradeSide := oppositeSide(p.Side)
		size := fmt.Sprint(p.PositionQuantity)
		if waitForSpread(r) {
			res := gmo.NewCloseAll(r, SYMBOL, tradeSide, "MARKET", "", size, "")
			if res != nil {
				ids = append(ids, res.Data)
			} else {
				logger("marketCloseBoth failed.Response was nil.")
			}
		}
	}
	return ids
}

func newTicker(r *gmo.ReqHandler) *gmo.TickerData {
	if res := gmo.NewTicker(r, SYMBOL); res == nil {
		return nil
	} else {
		return res.Extract(SYMBOL)
	}
}

// spread1が閾値以下になるのを待つ
// cnt分待っても閾値以下にならなければfalseを返す
func waitForSpread(r *gmo.ReqHandler) bool {
	cnt := 100
	for i := 0; i < cnt; i++ {
		ti := newTicker(r)
		if ti != nil {
			sp := ti.Spread()
			if sp < SPREAD_THRESH {
				return true
			}
			time.Sleep(500 * time.Millisecond) //閾値以上なら0.5秒松
		}
	}
	logger("Spread time out...")
	return false
}

// gmo.Excecutionsが取得出来るまで待ち、lossGainを返す
func waitForLossGain(r *gmo.ReqHandler, oid string) float64 {
	cnt := 7
	sum := 0.0
	for i := 0; i < cnt; i++ {
		res := gmo.NewExecutions(r, oid, "")
		if res != nil {
			for _, d := range res.Data.List {
				lg := d.LossGain
				if flossg, err := strconv.ParseFloat(lg, 64); err == nil {
					sum += flossg
				}
			}
			return sum
			//lg := res.Data.List[0].LossGain
			//return lg
		}
		//nilなら2秒待つ
		time.Sleep(2 * time.Second)
	}
	logger("waitForLossGain timeout... returning 0.0")
	return sum
}

// 注文ステータスがEXECUTED（全量約定）まで待つ関数
func waitUntilExecuted(r *gmo.ReqHandler, oid string) bool {
	cnt := 10
	for i := 0; i < cnt; i++ {
		res := gmo.NewOrders(r, oid)
		if res != nil && res.Status == 0 && len(res.Data.List) > 0 {
			data := res.Data.List[0]
			if data.Status == "EXECUTED" {
				return true
			}
		}
		time.Sleep(3 * time.Second)
	}
	//所定の時間待っても全量約定しなかった場合はfalse
	return false
}

// closeIds分waitForLossGainを呼び出し、損益を合計して文字列で返す
func getLossGain(r *gmo.ReqHandler, closeIds []string) string {
	sum := 0.0
	for _, c := range closeIds {
		//未約定の注文が存在すると集計されないので、注文ステータスがEXECUTED（全量約定）になるまで待つ
		b := waitUntilExecuted(r, c)
		//注文ステータスがEXECUTEDにならなくても処理自体は進め、ログ出力。
		sum += waitForLossGain(r, c)
		if !b {
			logger("close order status is not EXECUTED but continued...")
		}
	}
	return fmt.Sprint(sum)
}

// 保有positionのサマリから合計損益と保有量を返す
func getLossGainAndSizeFromPos(posList []gmo.Summary) (string, string) {
	lg := 0.0
	size := 0.0
	for _, p := range posList {
		lg += p.LossGain
		size += p.PositionQuantity
	}
	return fmt.Sprint(lg), fmt.Sprint(size)
}

// p:（平均）購入価格
// v:現在価格
// ratio:利益確定レート
// side:"BUY"or"SELL"
func isProfFilled(p, v, ratio float64, side string) bool {
	if side == "BUY" {
		return (v-p)/p >= ratio
	} else {
		return (p-v)/p >= ratio
	}
}

// p:（平均）購入価格
// v:現在価格
// ratio:損切レート
// side:"BUY"or"SELL"
func isLossFilled(p, v, ratio float64, side string) bool {
	if side == "BUY" {
		return (v-p)/p <= ratio
	} else {
		return (p-v)/p <= ratio
	}
}

// gmoのサマリを操作して利確する関数
// v:現在価格
// ratio:利確レシオ
func doProf(r *gmo.ReqHandler, pos []gmo.Summary, v, ratio float64) string {
	id := ""
	for _, p := range pos {
		avg := p.AveragePositionRate
		if isProfFilled(avg, v, ratio, p.Side) {
			logger(fmt.Sprintf("take profit. prof:%v", p.LossGain))
			_id := marketCloseSide(r, &p)
			if len(_id) > 0 {
				id = _id
			}
		}
	}
	return id
}

// 損切する関数
func doLoss(r *gmo.ReqHandler, pos []gmo.Summary, v, ratio float64) string {
	id := ""
	for _, p := range pos {
		avg := p.AveragePositionRate
		if isLossFilled(avg, v, ratio, p.Side) {
			logger(fmt.Sprintf("losscut. loss:%v", p.LossGain))
			_id := marketCloseSide(r, &p)
			if len(_id) > 0 {
				id = _id
			}
		}
	}
	return id
}

// c1,c2で長さ1以上のほうを返す
// 両方長さ1以上ならc1を返す
// 両方長さ０なら空文字を返す
func getCloseId(c1, c2 string) string {
	if len(c1) > 0 {
		return c1
	}
	if len(c2) > 0 {
		return c2
	}
	return ""
}

func genTweetText(prof, totalProf, valuation, posSize string) string {
	tags := "#BTC #Bitcoin"
	txt := "[" + getNow() + "]" + "\n" //[2022-4-5 23:00]
	txt += "🏄" + BOTNAME + VER + "🏄" + "\n"
	txt += "🚀確定損益:" + prof + "\n"
	txt += "🌝評価額:" + valuation + "\n"
	txt += "🌒保有量:" + posSize + "\n"
	txt += "🌜総利益 :" + totalProf + "\n"
	txt += tags
	return txt
}

// error時のツイートmessage
func genErrorTweetText(tprof string) string {
	tags := "#BTC #Bitcoin"
	txt := "[" + getNow() + "]" + "\n" //[2022-4-5 23:00]
	txt += "🏄" + BOTNAME + VER + "🏄" + "\n"
	txt += "🌜総利益 :" + tprof + "\n"
	txt += "\nエラーで今回は処理できませんでした\n"
	txt += tags
	return txt
}

// main処理
// 一定間隔のバッチで実行することを想定
// 基本１ポジションのみ持っていることを想定
// 複数あった場合にはエラーにはならないが、グラフとか総利益とかの値は保証されない
func live() {
	//**************************************************
	//初期化
	//**************************************************
	req := gmo.InitGMO("./conf.json") //リクエストハンドラ初期化

	_dCnt := 80       //ろうそく足の数
	dCnt := _dCnt + 1 //_dCnt分のろうそく足を評価用、直近を現在価格とするため+1する

	// sqlite3 開く
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	candles := newCandles(req, SYMBOL, TRADE_INTERVAL, dCnt) //ロウソク足取得
	summaries := gmo.NewSummary(req, SYMBOL)

	twitter := gotweet.NewTwitter("./twitter_conf.json")

	if candles != nil && summaries != nil && summaries.Status == 0 {
		//**************************************************
		//必要なデータが取得できた場合
		//**************************************************

		posList := summaries.Data.List //サマリデータのリスト
		//openTime
		otime := candles.OpenTime[len(candles.OpenTime)-1]
		//最後のindex -1でなく-2。直近はまだロウソク足形成中のため除外
		lasti := len(candles.Close) - 2
		//直近の(確定しているロウソク足の)終値
		latest := candles.Close[lasti]
		//inf初期化。Wrapも付ける。
		inf := minmax.NewInf(candles.High[:lasti], candles.Low[:lasti]).AddWrap(latest)
		//closeidのリスト
		closeIds := []string{}
		//openid
		openId := ""
		//判定用変数
		dec := ""

		//**********************************************************
		//ブレイクスルーなら持ちポジを全決済。
		//持ちポジが無い場合、posListは空スライスなのでそのままmarketClose呼ぶ
		//**********************************************************
		dec = breakThrough(latest, inf)
		if dec != "" {
			//　breakthroughと逆向きのpositionを持っている場合、決済する
			if len(posList) > 0 && posList[0].Side != dec {
				ids := marketCloseBoth(req, posList)
				closeIds = append(closeIds, ids...)
				if len(ids) > 0 {
					logger(fmt.Sprintf("breakthrough:latest:%.f max:%.f min:%.f", latest, inf.Maxv, inf.Minv))
					// 決済のDB更新はもっと下でやる
				}
			}
		}

		//*******************************************************
		//losscut,profit
		//doLoss,doProf両方実行される想定は無いのでcid１つにしてる
		//doLoss,doProfのいずれかの戻り値を設定するためにgetCloseIdを呼ぶ
		//*******************************************************
		c1 := doLoss(req, posList, latest, LOSS_RATIO)
		c2 := doProf(req, posList, latest, PROF_RATIO)
		cid := getCloseId(c1, c2)
		if len(cid) > 0 {
			closeIds = append(closeIds, cid)
			// 決済のDB更新はもっと下でやる

		}

		//*******************************************************
		// decが設定されていなかったらfibで設定
		// 24時間以上ポジションを持っている場合、損益レートを下げて利確。
		//*******************************************************
		if dec == "" {
			// positionありの場合、利確損切ラインを-1%,1%に縮めて再判定
			// 上で決済されている場合も、`あり`判定になるのに留意。
			// doProf,doLossでエラーになる可能性あるが、panicせず空文字を返すはずなので
			// そのままにしとく。
			if len(posList) > 0 {
				openPos := gmo.NewPositions(req, SYMBOL, "", "")
				if openPos != nil && len(openPos.Data.List) > 0 {
					// gmo PositionsのTimestamp "2022-02-12T12:12:12:12.011Z"のフォーマット
					openDateStr := openPos.Data.List[0].Timestamp
					// 日本時間にした上でtime.Time型に変換
					openDateTime := convUTCStringToJSTStamp(openDateStr)
					// 直近ろうそく足のopenTimeをtime.Time型に。
					ot := time.Unix(int64(otime), 0)

					hours := ot.Sub(openDateTime).Hours()

					// positionを持ってから24時間以上経過してたら短縮利確
					if hours >= 24 {
						c1 := doLoss(req, posList, latest, -0.01)
						c2 := doProf(req, posList, latest, 0.01)
						// 後処理
						cid := getCloseId(c1, c2)
						if len(cid) > 0 {
							closeIds = append(closeIds, cid)
							// 決済のDB更新はもっと下でやる
						}
					}
				}
			}
			dec = fib(inf)
		}

		//*******************************************************
		//新規取引処理
		//*******************************************************
		if dec != "" {
			//保有ポジなしもしくは上で決済されている場合は新規取引
			if len(posList) == 0 || len(closeIds) > 0 {
				openId = marketOpen(req, dec, TSIZE)

			}
		}

		//保有ポジありかつ未決済ならメッセージ出力
		if len(posList) > 0 && len(closeIds) == 0 {
			fmt.Println("Gopher already has a position!")
		}

		//*******************************************************
		// DB更新系
		//*******************************************************
		var totalP, fixedProf string
		//決済している場合、closeIdsをDBに挿入
		if len(closeIds) > 0 {
			insertCloseIds(db, closeIds, uint64(otime))
		}

		//決済している場合、取引履歴と損益をDBに挿入
		if len(closeIds) > 0 {
			fixedProf = getLossGain(req, closeIds) //今回の確定損益
			// fixedProfをint64に変換
			fixedProfInt, err := strconv.ParseInt(fixedProf, 10, 64)
			if err != nil {
				logger("Could not parse fixedProf to int: " + fixedProf)
			}
			// dbから直近の累計損益を取得
			recentBalance := selectRecentBalance(db)
			// 累計損益を更新
			totalPInt := int(fixedProfInt) + recentBalance
			// db「更新」。
			// 「既存ポジション決済」→「新規取引」が同一フレームで実行される可能性があるため、
			// db「挿入」(insertHistory)より前に実行すること。
			// 新規取引のdb挿入が先に実行されると、それが更新されてしまうため。
			updateHistory(db, &History{
				CloseTime: uint64(otime),
				// 前から決済の向きはposList[0].Sideから取得していたので活かす。
				// 決済時のポジションなので、逆向きにする。
				// 向きが異なるポジションがあった場合は最初のポジション向きでみなす。
				CloseSide:  oppositeSide(posList[0].Side),
				ClosePrice: uint64(latest),
				CloseAmt:   TSIZE_F,
				Profit:     int(fixedProfInt),
				Balance:    totalPInt,
			})
			// 後で使うように文字列としてセット
			totalP = string(fmt.Sprint(totalPInt))
		}

		// 新規取引をしている場合、DBに挿入
		// 「既存ポジション決済」→「新規取引」が同一フレームで実行される可能性があるため、
		// db更新（updateHistory）の後に実行する。
		if len(openId) > 0 {
			insertHistory(db, &History{
				OpenTime:  uint64(otime),
				OpenSide:  dec,
				OpenPrice: uint64(latest),
				OpenAmt:   TSIZE_F,
			})
			insertOpenId(db, openId, uint64(otime))
		}

		//グラフ用のロウソク足を取得して出力
		if cdGraph := newCandles(req, SYMBOL, TRADE_INTERVAL, 500); cdGraph != nil {
			AddCandleData(cdGraph) //グラフ用　ロウソク足出力
		}

		//*******************************************************
		//tweet
		//*******************************************************
		var valuation, posSize string

		if len(closeIds) == 0 && len(posList) > 0 {
			//保有positionがあり、今回決済されていない場合、評価額と保有量をセット
			valuation, posSize = getLossGainAndSizeFromPos(posList)
		} else if len(openId) > 0 {
			//今回新規取引している場合、新たにサマリを取得して設定
			fmt.Println("[test] in len(openId)>0")
			time.Sleep(10 * time.Millisecond)
			if nSum := gmo.NewSummary(req, SYMBOL); nSum != nil && nSum.Status == 0 {
				fmt.Println("[test] in if nSum:=gmo.NewSummary...")
				valuation, posSize = getLossGainAndSizeFromPos(nSum.Data.List)
			} else {
				fmt.Println("[test] NewSummary failed?")
			}
		}
		//tweet用テキスト生成
		tweetTxt := genTweetText(fixedProf, totalP, valuation, posSize)
		//グラフを画像として出力するpythonファイルを実行
		cmd := exec.Command(genPyCommand(), PYSCRIPT)
		if b, err := cmd.CombinedOutput(); err != nil {
			//err時は表示
			fmt.Println(err)
			fmt.Println(string(b))
		}

		twitter.Tweet(tweetTxt, IMG_PATH)
		fmt.Printf("latest:%.f,max:%.f,min:.%f,scale:%f,decision:%v\n", latest, inf.Maxv, inf.Minv, inf.Scaled, dec)

	} else {
		//**************************************************
		//データ取れなかった場合
		//**************************************************
		logger("could not get candles or summary response...")
		tProf := selectRecentBalance(db)
		twitter.Tweet(genErrorTweetText(fmt.Sprint(tProf)))

	}
}

func test() {
	bktest.Backtest()
}

func main() {
	/*live() or backtest()*/

	//global変数設定
	setGlobalVars()

	if TEST_MODE {
		test()
	} else {
		live()
	}
}
