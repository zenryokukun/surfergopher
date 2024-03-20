package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/zenryokukun/surfergopher/gmo"
)

func newCandles(r *gmo.ReqHandler, sym, itv string, cnt int) *gmo.CandlesData {
	layout := dateLayout(itv)        //itvに応じてYYYY or YYYYMMDDを取得
	day := time.Now().Format(layout) //現在時刻をYYYMMDD or YYYYにフォーマット

	data := &gmo.CandlesData{}
	i := 0 //日付制御用 一回の実行でcntを満たさない場合、-1ずつ減らして日付の減算に使う。

	for {
		if len(data.Close) > cnt {
			//cntより大きかったらBREAK!
			//fmt.Printf("data len:%v breaking...\n", len(data.Close))
			break
		}
		//api 実行
		tmpres := gmo.NewCandles(r, sym, itv, day)
		if tmpres == nil || tmpres.Status != 0 {
			//error時はnilを返す
			fmt.Printf("itv:%v,day:%v\n", itv, day)
			fmt.Printf("i:%v tmpres nil or status != 0\n", i)
			return nil
		}
		//CandlesDataに変換
		if tmpdata := tmpres.Extract(); tmpdata != nil {
			data.AddBefore(tmpdata) //既存データの前に入れる
		}

		i -= 1 //前日(年）分とするため-1
		if len(layout) == 8 {
			day = time.Now().AddDate(0, 0, i).Format(layout) //yyyymmddの場合は日付を-1
		} else if len(layout) == 4 {
			day = time.Now().AddDate(i, 0, 0).Format(layout) //yyyyの場合は年を-1
		}
		//fmt.Printf("next day:%v\n", day)
		time.Sleep(500 * time.Millisecond) //0.5秒待つ!
	}
	//おしりからcnt分だけを返す
	l := len(data.Close)
	data.Slice(l-cnt, l)
	return data
}

// GetCandlesのitvに応じた日付レイアウトを返す関数
// goでは"2006"-> YYYY, "01" -> MM, "02" -> DD　のフォーマットにされる仕様
func dateLayout(itv string) string {
	var layout string
	if itv == "4hour" || itv == "1day" || itv == "1week" || itv == "1month" {
		layout = "2006" //YYYYにフォーマットされる
	} else {
		layout = "20060102" //YYYYMMDDにフォーマットされる
	}
	return layout
}

// "2022-08-20T00:01:02.604Z"のフォーマットの文字列を
// 日本時間に変換した上でtime.Time型に変換。
// dstrはgmoのPositionsのTimestampを想定
// 取引時は9:01のように、1分の時に処理を回しているので0分になおしてる。
// 直近のロウソク足の時間と、openした時の時間を計算するために使う想定
func convUTCStringToJSTStamp(dstr string) time.Time {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	t, _ := time.Parse(time.RFC3339, dstr)
	jt := t.In(loc)
	nt := time.Date(jt.Year(), jt.Month(), jt.Day(), jt.Hour(), 0, 0, 0, loc)
	return nt
}

func getNow() string {
	now := time.Now().Format("2006-01-02 15:04")
	return now
}

func logger(msg string) {
	now := time.Now().Format(time.RFC3339)
	fmt.Printf("%v,%v\n", now, msg)
}

func oppositeSide(side string) string {
	if side == "BUY" {
		return "SELL"
	}
	if side == "SELL" {
		return "BUY"
	}
	return ""
}

// Linux,Windowsによってコマンドが違うのでここで解決する
func genPyCommand() string {
	//"windows" or "linux"
	switch runtime.GOOS {
	case "windows":
		return "python"
	case "linux":
		return "python3"
	default:
		return ""
	}
}

// ロウソク足をファイルに出力
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
