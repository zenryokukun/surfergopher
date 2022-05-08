package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	_ "os/exec"
	"strconv"
	"time"

	"github.com/zenryokukun/surfergopher/bktest"
	"github.com/zenryokukun/surfergopher/gmo"
	"github.com/zenryokukun/surfergopher/minmax"
)

const (
	TEST_MODE = false //本番かテストモードか
	//SYMBOL         = "BTC_JPY"
	GLOBAL_FPATH   = "./globals.json"
	CLOSEDID_FPATH = "./data/closedposlist.txt" //closeしたorderIdのリスト
	TPROF_FPATH    = "./data/totalprof.txt"     //総利益を保管しておくファイル
	//SPREAD_THRESH  = 1500.0                     //許容するスプレッド
	//TSIZE          = "0.1"                      //取引量
	BOTNAME  = "Surfer Gopher" //botの名前
	VER      = "@v1.0"         //botのversion
	PYSCRIPT = "./py/chart.py" //pythonスクリプト
	IMG_PATH = "./py/tweet.png"
)

//環境によって分けるためグローバル変数にした
var (
	SYMBOL         string
	TRADE_INTERVAL string
	TSIZE          string
	PROF_RATIO     float64
	LOSS_RATIO     float64
	SPREAD_THRESH  float64
)

//グローバル変数をファイルから設定
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
}

//Returns orderId
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

//Returns closeordeId
//
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

//Returns slice of closeOrderIds
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

//spread1が閾値以下になるのを待つ
//cnt分待っても閾値以下にならなければfalseを返す
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

//gmo.Excecutionsが取得出来るまで待ち、lossGainを返す
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

//注文ステータスがEXECUTED（全量約定）まで待つ関数
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

//closeIds分waitForLossGainを呼び出し、損益を合計して文字列で返す
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

//保有positionのサマリから合計損益と保有量を返す
func getLossGainAndSizeFromPos(posList []gmo.Summary) (string, string) {
	lg := 0.0
	size := 0.0
	for _, p := range posList {
		lg += p.LossGain
		size += p.PositionQuantity
	}
	return fmt.Sprint(lg), fmt.Sprint(size)
}

//総利益を更新しファイルに出力。更新した総利益を返す。
//prof -> 今回利益
func updateTotalProf(prof string) string {
	//string -> int
	profi, err := strconv.ParseInt(prof, 10, 32)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	//ファイルから総利益を取得
	b, err := os.ReadFile(TPROF_FPATH)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	//総利益をstring -> intに変換
	tprof, err := strconv.ParseInt(string(b), 10, 32)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	//総利益を更新し、int -> stringに変換
	newTProfStr := fmt.Sprint(tprof + profi)

	//新総利益をファイルに出力
	f, err := os.Create(TPROF_FPATH)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer f.Close()
	f.Write([]byte(newTProfStr))

	return newTProfStr
}

//総利益をファイルから取得する関数
func getTotalProf() string {
	b, err := os.ReadFile(TPROF_FPATH)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)

}

//ファイルにcloseしたorderIdを追記
func writeCloseIds(closeIds []string) {
	f, err := os.OpenFile(CLOSEDID_FPATH, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	for _, v := range closeIds {
		f.Write([]byte(v + "\n"))
	}
}

//p:（平均）購入価格
//v:現在価格
//ratio:利益確定レート
//side:"BUY"or"SELL"
func isProfFilled(p, v, ratio float64, side string) bool {
	if side == "BUY" {
		return (v-p)/p >= ratio
	} else {
		return (p-v)/p >= ratio
	}
}

//p:（平均）購入価格
//v:現在価格
//ratio:損切レート
//side:"BUY"or"SELL"
func isLossFilled(p, v, ratio float64, side string) bool {
	if side == "BUY" {
		return (v-p)/p <= ratio
	} else {
		return (p-v)/p <= ratio
	}
}

//gmoのサマリを操作して利確する関数
//v:現在価格
//ratio:利確レシオ
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

//損切する関数
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

//c1,c2で長さ1以上のほうを返す
//両方長さ1以上ならc1を返す
//両方長さ０なら空文字を返す
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

//error時のツイートmessage
func genErrorTweetText() string {
	tprof := getTotalProf()
	tags := "#BTC #Bitcoin"
	txt := "[" + getNow() + "]" + "\n" //[2022-4-5 23:00]
	txt += "🏄" + BOTNAME + VER + "🏄" + "\n"
	txt += "🌜総利益 :" + tprof + "\n"
	txt += "\nエラーで今回は処理できませんでした\n"
	txt += tags
	return txt
}

func live() {
	//**************************************************
	//初期化
	//**************************************************
	req := gmo.InitGMO("./conf.json") //リクエストハンドラ初期化
	//titv := "4hour"                   //trade interval
	//profR := 0.05                     //profit ratio
	//lossR := -0.05                    //losscut ratio
	_dCnt := 80       //ろうそく足の数
	dCnt := _dCnt + 1 //_dCnt分のろうそく足を評価用、直近を現在価格とするため+1する

	candles := newCandles(req, TRADE_INTERVAL, dCnt) //ロウソク足取得
	summaries := gmo.NewSummary(req, SYMBOL)

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
		//取引履歴用データ
		histo := &History{}
		//判定用変数
		dec := ""

		//**********************************************************
		//ブレイクスルーなら持ちポジを全決済。
		//持ちポジが無い場合、posListは空スライスなのでそのままmarketClose呼ぶ
		//**********************************************************
		dec = breakThrough(latest, inf)
		if dec != "" {
			ids := marketCloseBoth(req, posList)
			closeIds = append(closeIds, ids...)
			if len(ids) > 0 {
				logger(fmt.Sprintf("breakthrough:latest:%.f max:%.f min:%.f", latest, inf.Maxv, inf.Minv))
				histo.addHistory(otime, latest, dec, "CLOSE")
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
		}

		//*******************************************************
		//decが設定されていなかったらfibで設定
		//*******************************************************
		if dec == "" {
			dec = fib(inf)
		}

		//*******************************************************
		//新規取引処理
		//*******************************************************
		if dec != "" {
			//保有ポジなしもしくは上で決済されている場合は新規取引
			if len(posList) == 0 || len(closeIds) > 0 {
				openId = marketOpen(req, dec, TSIZE)
				if len(openId) > 0 {
					histo.addHistory(otime, latest, dec, "OPEN")
				}
			}
		}

		//保有ポジありかつ未決済ならメッセージ出力
		if len(posList) > 0 && len(closeIds) == 0 {
			fmt.Println("Gopher already has a position!")
		}

		//*******************************************************
		//ファイル出力系
		//*******************************************************
		var totalP, fixedProf string
		//closeIdsをファイルにアペンド
		writeCloseIds(closeIds)
		//決済している場合は総利益を更新してファイルに出力。
		if len(closeIds) > 0 {
			fixedProf = getLossGain(req, closeIds) //今回の確定損益
			updateTotalProf(fixedProf)             //総利益のファイルを更新
		}
		//総利益をファイルから取得。今回決済していない場合でも使うので上のif分の外に記載
		//float64に変換して残高推移ファイルに出力
		totalP = getTotalProf()
		if totalPfloat, err := strconv.ParseFloat(totalP, 64); err == nil {
			AddBalance(otime, totalPfloat)
		}
		//グラフ用のロウソク足を取得して出力
		if cdGraph := newCandles(req, TRADE_INTERVAL, 200); cdGraph != nil {
			AddCandleData(cdGraph) //グラフ用　ロウソク足出力
		}
		//取引履歴を出力
		AddTradeHistory(histo)

		//*******************************************************
		//tweet
		//*******************************************************
		var valuation, posSize string
		if len(closeIds) == 0 && len(posList) > 0 {
			//保有positionがあり、今回決済されていない場合、評価額と保有量をセット
			valuation, posSize = getLossGainAndSizeFromPos(posList)
			//総利益に評価額を加減し、総利益ファイルを書き換える。
			//グラフに評価額を表示させるために追加した。
			totalPWithVal := addStringFloat(totalP, valuation)
			ReplaceBalance(totalPWithVal)
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

		NewTwitter().tweetImage(tweetTxt, IMG_PATH)

		fmt.Printf("latest:%.f,max:%.f,min:.%f,scale:%f,decision:%v\n", latest, inf.Maxv, inf.Minv, inf.Scaled, dec)

	} else {
		//**************************************************
		//データ取れなかった場合
		//**************************************************
		logger("could not get candles or summary response...")
		NewTwitter().tweet(genErrorTweetText(), nil)
	}
}

func test() {
	bktest.Backtest()
}

func main() {
	/*live() or backtest()*/

	//global変数設定
	setGlobalVars()
	//必要なファイルが無い場合はからファイル作成
	doesExist(
		//POSITION_FPATH,
		//ORDERID_FPATH,
		CLOSEDID_FPATH,
		TPROF_FPATH,
	)
	//graph.go内のパス
	doesExist(
		CDATA_FPATH,
		BDATA_FPATH,
		TRADE_FPATH,
	)

	if TEST_MODE {
		test()
	} else {
		live()
	}

}

/*
func live() {
	req := gmo.InitGMO("./conf.json") //リクエストハンドラ初期化
	titv := "4hour"                   //trade interval
	profR := 0.05                     //profit ratio
	lossR := -0.05                    //losscut ratio
	_dCnt := 80                       //ろうそく足の数
	dCnt := _dCnt + 1                 //_dCnt分のろうそく足を評価用、直近を現在価格とするため+1する

	//ロウソク足取得
	candles := newCandles(req, titv, dCnt)
	//openTime
	otime := candles.OpenTime[len(candles.OpenTime)-1]
	//最後のindex -1でなく-2。直近はまだロウソク足形成中のため除外
	lasti := len(candles.Close) - 2
	//直近の(確定しているロウソク足の)終値
	latest := candles.Close[lasti]
	//inf初期化。Wrapも付ける。
	inf := minmax.NewInf(candles.High[:lasti], candles.Low[:lasti]).AddWrap(latest)
	//botのポジションのみ取得。配列で返ってくるので注意。
	//pos := filterPosition(req)
	//建玉一覧
	summary := newSummary(req)
	hasPos := summary.Has() //position保有フラグ
	//action初期化
	act := NewAction(len(pos))
	//トレンドを入れる変数初期化。"BUY" or "SELL"
	dec := ""
	//このフレームでcloseしたときのorderId
	closeId := ""
	//このフレームでopenしたときのopenId
	openId := ""
	//このフレームで確定した利益
	fixedProf := "0"
	//このフレームでの取引を加味した総損益
	totalProf := "0"
	//取引情報
	histo := &History{}
	//直近価格が最大値を超えている場合は買い判定をセット。
	//売玉がある場合は決済もする。
	if latest > inf.Maxv {
		//ポジションを持っていなければopen-buy判定
		if hasPos == false {
			dec = "BUY"
		} else {
			//p := &pos[0] //配列なので[0]。複数の想定は無いが、その場合も[0]で処理
			if summary.Side == "SELL" {
				//CloseBuy
				//CLOSE ALL を要実装
				//closeId = marketClose(req, p, act)
				if len(closeId) > 0 {
					//marketCloseが成功していたら取引履歴を更新
					histo.addHistory(otime, latest, "BUY", "CLOSE")
				}
				dec = "BUY"
			}
		}
		logger(fmt.Sprintf("latest:%.f >> max:%.f", latest, inf.Maxv))

	} else if latest <= inf.Minv {
		//直近価格が最小値を下回っている場合は売り判定
		//買玉がある場合は決済もする。
		if hasPos == false {
			dec = "SELL"
		} else {
			//p := &pos[0]
			if summary.Side == "BUY" {
				//CloseSell
				closeId = marketClose(req, p, act)
				if len(closeId) > 0 {
					//marketCloseが成功していたら取引履歴を更新
					histo.addHistory(otime, latest, "SELL", "CLOSE")
				}
				dec = "SELL"
			}
		}
		logger(fmt.Sprintf("latest:%.f << max:%.f", latest, inf.Maxv))
	}

	//ポジションある時は利確・損切判定処理実施
	if len(pos) > 0 {
		p := &pos[0]
		if isLossFilled(summary, latest, lossR) {
			//Close
			closeId = marketClose(req, p, act)
			if len(closeId) > 0 {
				//marketCloseが成功していた場合
				s := oppositeSide(p.Side)                   //positionと逆のサイド
				histo.addHistory(otime, latest, s, "CLOSE") //履歴更新
				logger(fmt.Sprintf("profitFilled.latest:%.f posPrice:%.v", latest, p.Price))
			}
		}
		if isProfFilled(summary, latest, profR) {
			//close
			closeId = marketClose(req, p, act)
			if len(closeId) > 0 {
				s := oppositeSide(p.Side)                   //positionと逆のサイド
				histo.addHistory(otime, latest, s, "CLOSE") //履歴更新
				logger(fmt.Sprintf("lossFilled.latest:%.f posPrice:%.v", latest, p.Price))
			}
		}
	}

	//decが未設定の場合、fiboLevelに応じて設定
	if dec == "" {

		lvl := fibo.Level(inf.Scaled)

		if lvl >= 5 {
			if inf.Which == "B" {
				dec = "BUY"
			} else if inf.Which == "T" {
				dec = "SELL"
			}
			logger(fmt.Sprintf("fibLvl>=5. recent:%v scaled:%v decision:%v", inf.Which, inf.Scaled, dec))
		}

		if lvl <= 1 {
			if inf.Which == "B" {
				dec = "BUY"
			} else if inf.Which == "T" {
				dec = "SELL"
			}
			logger(fmt.Sprintf("fibLvl<=1. recent:%v scaled:%v decision:%v", inf.Which, inf.Scaled, dec))
		}

		//added
		if lvl == 4 {
			if inf.Which == "B" {
				dec = "BUY"
			} else if inf.Which == "T" {
				dec = "SELL"
			}
		}

	}

	//取引判断ありかつ保有positionなしならopen
	npos := filterPosition(req) //上で決済している可能性もあるため、再度保有positionを取得
	if len(npos) > 0 {
		logger("has pos. not trading in this frame...")
	}
	if dec != "" && len(npos) == 0 {
		//open
		openId = marketOpen(req, dec, TSIZE, act)
		//marketOpenが成功していた時の処理
		if len(openId) > 0 {
			histo.addHistory(otime, latest, dec, "OPEN")
		}
	}

	fmt.Printf("latest:%.f,max:%.f,min:.%f,scale:%f,decision:%v\n", latest, inf.Maxv, inf.Minv, inf.Scaled, dec)

	//closeIdが設定されている場合、総利益を更新
	if len(closeId) > 0 {
		fixedProf = waitForLossGain(req, closeId)
		totalProf = updateTotalProf(fixedProf)
	}

	//************************************************************
	//以下はtweet用情報
	//************************************************************
	valuation := "0"
	posSize := "0"
	if len(closeId) == 0 {
		totalProf = getTotalProf()
		if len(pos) > 0 {
			valuation = pos[0].LossGain
			posSize = pos[0].Size
		}
	}

	if len(openId) > 0 {
		posSize = TSIZE
	}

	txt := genTweetText(fixedProf, totalProf, valuation, posSize)

	//グラフ用データ設定
	tpFloat, err := strconv.ParseFloat(totalProf, 64)
	if err != nil {
		fmt.Println(err)
	}

	//tweet画像用のファイル出力
	cdGraph := newCandles(req, titv, 200)
	AddTradeHistory(histo)
	AddBalance(otime, tpFloat)
	AddCandleData(cdGraph)
	//画像生成python script
	cmd := exec.Command(genPyCommand(), PYSCRIPT)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(b))
	}
	NewTwitter().tweetImage(txt, IMG_PATH)

}
*/
