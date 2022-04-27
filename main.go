package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/zenryokukun/surfergopher/fibo"
	"github.com/zenryokukun/surfergopher/gmo"
	"github.com/zenryokukun/surfergopher/minmax"
)

const (
	SYMBOL         = "BTC_JPY"
	POSITION_FPATH = "./data/poslist.txt"       //このbotのポジションを書き込むファイルパス
	ORDERID_FPATH  = "./data/orderid.txt"       //直近のorderIdを保存
	CLOSEDID_FPATH = "./data/closedposlist.txt" //closeしたorderIdのリスト
	TPROF_FPATH    = "./data/totalprof.txt"     //総利益を保管しておくファイル
	SPREAD_THRESH  = 1200.0                     //許容するスプレッド
	TSIZE          = "0.1"                      //取引量
	BOTNAME        = "Surfer Gopher"            //botの名前
	VER            = "@v1.0"                    //botのversion
	PYSCRIPT       = "./py/chart.py"            //pythonスクリプト
	IMG_PATH       = "./py/tweet.png"
)

type (
	action struct {
		has   bool
		close bool
		open  bool
	}
)

func (a *action) closed() {
	a.close = true
}
func (a *action) opened() {
	a.open = true
}

func NewAction(plen int) *action {
	var has bool = true
	if plen == 0 {
		has = false
	}
	return &action{has: has, close: false, open: false}
}

func _marketOpen(r *gmo.ReqHandler, side, size string) *gmo.OpenRes {
	res := gmo.NewOpenOrder(r, SYMBOL, side, "MARKET", "", size, "", "")
	return res
}

//Returns orderId
func marketOpen(r *gmo.ReqHandler, side, size string, act *action) string {
	if waitForSpread(r) {
		res := _marketOpen(r, side, size)
		if res != nil {
			//fileにposition追加
			pid := orderIdToPosId(r, res.Data)
			addPosition(pid)
			//actにopenフラグ追加
			act.opened()
			return res.Data
		} else {
			logger("marketOpen failed.Response was nil.")
		}
	}
	return ""
}

func _marketClose(r *gmo.ReqHandler, side, size, posId string) *gmo.CloseRes {
	res := gmo.NewCloseOrder(r, SYMBOL, side, "MARKET", "", posId, size, "")
	return res
}

//Returns orderId
func marketClose(r *gmo.ReqHandler, pos *gmo.Positions, act *action) string {
	var side string
	if pos.Side == "BUY" {
		side = "SELL"
	} else {
		side = "BUY"
	}
	if waitForSpread(r) {
		res := _marketClose(r, side, pos.Size, fmt.Sprint(pos.PositionId))
		if res != nil {
			//fileからポジション削除
			pid := orderIdToPosId(r, res.Data) //orderIdに対応するpositionIdを取得
			removePosition(pid)
			//actにcloseフラグつける
			act.closed()
			return res.Data
		} else {
			logger("marketClose failed.Response was nil.")
		}
	}
	return ""
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

func orderIdToPosId(r *gmo.ReqHandler, orderId string) uint32 {
	res := gmo.NewLatestExecutions(r, SYMBOL, "", "")
	pid := res.FindByOrderId(orderId)
	return strToUint32(pid)
}

//ポジション一覧をapiから取得
func positions(r *gmo.ReqHandler) []gmo.Positions {
	res := gmo.NewPositions(r, "BTC_JPY", "", "")
	if res == nil {
		return nil
	}
	return res.Data.List
}

//spread1が閾値以下になるのを待つ
//cnt分待っても閾値以下にならなければfalseを返す
func waitForSpread(r *gmo.ReqHandler) bool {
	cnt := 7
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
func waitForLossGain(r *gmo.ReqHandler, oid string) string {
	cnt := 7
	for i := 0; i < cnt; i++ {
		res := gmo.NewExecutions(r, oid, "")
		if res != nil {
			lg := res.Data.List[0].LossGain
			return lg
		}
		//nilなら2秒待つ
		time.Sleep(2 * time.Second)
	}
	return ""
}

//BOTのポジションをファイルから取得
func myPositionId() []uint32 {
	//open file READONLY
	f, err := os.Open(POSITION_FPATH)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer f.Close()

	ids := []uint32{}

	//read line by line,append to ids.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		id := scanner.Text()
		uintid, err := strconv.ParseUint(id, 10, 32) //unit64
		if err != nil {
			return nil
		}
		ids = append(ids, uint32(uintid))
	}

	//return nil if len is 0
	if len(ids) == 0 {
		return nil
	}
	return ids
}

//apiのポジションからbotのポジションのみを取得
func filterPosition(r *gmo.ReqHandler) []gmo.Positions {
	mp := myPositionId() //all positions from api
	gp := positions(r)   //bot positions
	botPos := []gmo.Positions{}
	for _, mv := range mp {
		for _, gv := range gp {
			if mv == gv.PositionId {
				botPos = append(botPos, gv)
			}
		}
	}
	return botPos
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
func addClosedPos(myId string) {
	f, err := os.OpenFile(CLOSEDID_FPATH, os.O_APPEND, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte(myId + "\n"))
}

//fileにmyIdを追加（os.O_APPEND)
func addPosition(myId uint32) {
	f, err := os.OpenFile(POSITION_FPATH, os.O_APPEND, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	strId := fmt.Sprint(myId) + "\n"
	f.Write([]byte(strId))
}

//ファイルからmyIdのポジションを削除する
func removePosition(myId uint32) {
	myPos := myPositionId() //ファイルからid取得
	newPos := []uint32{}    //myId以外のidを入れるスライス
	exists := false         //一致するものが存在するかフラグ

	for _, pid := range myPos {
		if pid != myId {
			newPos = append(newPos, pid) //不一致ならnewPosにappend
		} else {
			exists = true //一致した場合はフラグをセット
		}
	}

	if exists == false {
		return //一致するものがなかったら戻る
	}

	if f, err := os.Create(POSITION_FPATH); err == nil {
		if len(newPos) == 0 {
			//出力するものが何もない場合（最後の１つを削除）、空文字を出力
			f.Write([]byte(""))
		} else {
			//ある場合はuint32->stringに変換して改行文字つけて出力
			txt := ""
			for _, v := range newPos {
				txt += fmt.Sprint(v) + "\n"
			}
			//txt = txt[:len(txt)-1] //最後の改行文字は削除
			f.Write([]byte(txt))
		}
	} else {
		fmt.Println(err)
	}
}

//posPrice ->*gmo.Positions
//v   ->現在価格
//ratio    ->利益確定レート
func isProfFilled(pos *gmo.Positions, v float64, ratio float64) bool {
	oprice, err := strconv.ParseFloat(pos.Price, 64)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if pos.Side == "BUY" {
		return (v-oprice)/oprice >= ratio
	} else {
		return (oprice-v)/oprice >= ratio
	}
}

//posPrice ->*gmo.Positions
//v   ->現在価格
//ratio    ->損切レート
func isLossFilled(pos *gmo.Positions, v float64, ratio float64) bool {
	oprice, err := strconv.ParseFloat(pos.Price, 64)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if pos.Side == "BUY" {
		return (v-oprice)/oprice <= ratio
	} else {
		return (oprice-v)/oprice <= ratio
	}
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
	pos := filterPosition(req)
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
		if len(pos) == 0 {
			dec = "BUY"
		} else {
			p := &pos[0] //配列なので[0]。複数の想定は無いが、その場合も[0]で処理
			if p.Side == "SELL" {
				//CloseBuy
				closeId = marketClose(req, p, act)
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
		if len(pos) == 0 {
			dec = "SELL"
		} else {
			p := &pos[0]
			if p.Side == "BUY" {
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
		if isLossFilled(p, latest, lossR) {
			//Close
			closeId = marketClose(req, p, act)
			if len(closeId) > 0 {
				//marketCloseが成功していた場合
				s := oppositeSide(p.Side)                   //positionと逆のサイド
				histo.addHistory(otime, latest, s, "CLOSE") //履歴更新
				logger(fmt.Sprintf("profitFilled.latest:%.f posPrice:%.v", latest, p.Price))
			}
		}
		if isProfFilled(p, latest, profR) {
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
	}

	//取引判断ありかつ保有positionなしならopen
	if dec != "" && len(pos) == 0 {
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

	/************************************************************/
	//以下はtweet用情報
	/************************************************************/
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

	/*グラフ用データ設定*/
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

func main() {
	/*live() or backtest()*/
	//必要なファイルが無い場合はからファイル作成
	doesExist(
		POSITION_FPATH,
		ORDERID_FPATH,
		CLOSEDID_FPATH,
		TPROF_FPATH,
	)
	//graph.go内のパス
	doesExist(
		CDATA_FPATH,
		BDATA_FPATH,
		TRADE_FPATH,
	)

	live()
}
