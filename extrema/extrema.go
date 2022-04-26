package extrema

type (
	Extrema struct {
		Val  []float64
		Time []int
	}
)

func (e *Extrema) Add(val float64, utime int) {
	e.Val = append(e.Val, val)
	e.Time = append(e.Time, utime)
}

func SumFloat(y []float64) float64 {
	sum := 0.0
	for _, v := range y {
		sum += v
	}
	return sum
}

func AvgFloat(y []float64) float64 {
	return SumFloat(y) / float64(len(y))
}

func Search(x []int, y []float64, ratio float64) (maxima *Extrema, minima *Extrema) {
	asc := true //始めは高値検索
	index := 0  //検索位置
	ctrl := 0   //制御用。高値検索、底値検索が連続して見つからない場合を把握するために使う。
	maxima = &Extrema{}
	minima = &Extrema{}

	//最後まで検索する
	for index < len(y) {
		step := 0 //高値底値のindex.
		if asc {
			step = searchMax(y, index, ratio) //高値検索
		} else {
			step = searchMin(y, index, ratio) //底値検索
		}
		if step == -1 || step == 0 {
			//見つからないの処理
			ctrl++
			if ctrl == 2 {
				break //2連続で見つからない場合は処理終了
			}

		} else {
			//見つかった場合の処理
			index += step //indexを高値底値の位置にする
			if index < len(y) {
				minormax(maxima, minima, asc).Add(y[index], x[index]) //maxima or minimaに値をセット
			}
			ctrl = 0 //制御用変数をリセット
		}
		asc = !asc //高値底値を反転
	}
	return maxima, minima
}

//最大値検索 yは検査配列 startは検索開始位置
func searchMax(y []float64, start int, ratio float64) int {
	notfound := -1  //見つからない場合は-1を返す
	max := y[start] //startの値で初期化
	pos := 0        //オフセット
	for i, v := range y[start:] {
		if v >= max {
			//最大値を更新した場合　高堰とオフセットを更新
			max = v
			pos = i
		} else {
			if isExtrema(max, v, ratio, true) {
				//最大値より低い場合、ratio以上の下落があれば高値確定。オフセットを返して終了
				return pos
			}
		}
	}
	return notfound
}

//searchMaxの底値version.
func searchMin(y []float64, start int, ratio float64) int {
	notfound := -1
	min := y[start]
	pos := 0
	for i, v := range y[start:] {
		if v <= min {
			min = v
			pos = i
		} else {
			if isExtrema(min, v, ratio, false) {
				return pos
			}
		}
	}
	return notfound
}

//極値をFIXさせるかチェックする。
//checkVal:極値候補　currentVal:現在値
//topがtrueなら最大値、falseなら最小値チェック。
func isExtrema(checkVal, currentVal, ratio float64, top bool) bool {
	if top {
		//（最大値-現在値）/最大値がraito以上になれば最大値FIX
		return (checkVal-currentVal)/checkVal >= ratio
	} else {
		//(現在値-最小値)/最大値がratio以上になれば最小値FIX
		return (currentVal-checkVal)/checkVal >= ratio
	}
}

func minormax(max *Extrema, min *Extrema, asc bool) *Extrema {
	if asc {
		return max
	} else {
		return min
	}
}
