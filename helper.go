package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/zenryokukun/surfergopher/gmo"
)

func newCandles(r *gmo.ReqHandler, itv string, cnt int) *gmo.CandlesData {
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
		tmpres := gmo.NewCandles(r, "BTC_JPY", itv, day)
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

//GetCandlesのitvに応じた日付レイアウトを返す関数
//goでは"2006"-> YYYY, "01" -> MM, "02" -> DD　のフォーマットにされる仕様
func dateLayout(itv string) string {
	var layout string
	if itv == "4hour" || itv == "1day" || itv == "1week" || itv == "1month" {
		layout = "2006" //YYYYにフォーマットされる
	} else {
		layout = "20060102" //YYYYMMDDにフォーマットされる
	}
	return layout
}

func toTime(ut int64) time.Time {
	return time.Unix(ut, 0)
}

func strToUint32(s string) uint32 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return uint32(i)
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

//Linux,Windowsによってコマンドが違うのでここで解決する
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

//パラメタのファイルが存在しなければ作る
func doesExist(s ...string) {
	for _, v := range s {
		_, err := os.Stat(v)
		if err != nil && os.IsNotExist(err) {
			f, err := os.Create(v)
			if err != nil {
				fmt.Println(err)
			}
			//TPROF_FPATHなら"0"を入れておく
			if v == TPROF_FPATH {
				f.Write([]byte("0"))
			}
			f.Close()
		}
	}
}

//文字列→小数点に変換して足し算して返す関数
//["1.1","2.2"] -> 3.3
func addStringFloat(f ...string) float64 {
	sum := 0.0
	for _, v := range f {
		if vfloat, err := strconv.ParseFloat(v, 64); err == nil {
			sum += vfloat
		}
	}
	return sum
}
