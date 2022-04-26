package minmax

type (
	//検索元全量データ。高値、底値、終値を想定
	minmax struct {
		high []float64
		low  []float64
		//close  []float64 //使ってないかも、、、
		//tIndex []int     //使ってないかも、、、
	}

	Inf struct {
		Maxi   int
		Maxv   float64
		Mini   int
		Minv   float64
		Which  string
		Scaled float64
	}
)

//Exposed Constructor of minmax.
func New(high, low []float64) *minmax {
	if len(high) != len(low) {
		return nil
	}
	mm := &minmax{
		high: high, low: low,
	}
	return mm
}

//Exposed Constructor of inf
func NewInf(high, low []float64) *Inf {
	mm := New(high, low)
	mini, minv := mm.SearchMin(0)
	maxi, maxv := mm.SearchMax(0)
	which := mm.Recent(0)
	return &Inf{
		Maxi: maxi, Mini: mini,
		Maxv: maxv, Minv: minv,
		Which: which,
	}
}

//Method:Returns index and value of the minimum value.
//start -> 検索開始位置
func (mm *minmax) SearchMin(start int) (int, float64) {
	mv := mm.low[start]
	mi := start
	for i, v := range mm.low[start:] {
		if v <= mv {
			mi = i
			mv = v
		}
	}
	return mi, mv
}

//Method:Returns index and value of the maximum value.
//start -> 検索開始位置
func (mm *minmax) SearchMax(start int) (int, float64) {
	mv := mm.high[start]
	mi := start
	for i, v := range mm.high[start:] {
		if v >= mv {
			mi = i
			mv = v
		}
	}
	return mi, mv
}

//returns maxIndex,minIndex. Values are not returned.
func (mm *minmax) SearchMinMax(start int) (maxi int, mini int) {
	maxi, _ = mm.SearchMax(start)
	mini, _ = mm.SearchMin(start)
	return
}

//直近が高値か底値を返す
//"T" -> 高値　"B" -> 底値
func (mm *minmax) Recent(start int) string {
	maxi, mini := mm.SearchMinMax(start)
	if maxi > mini {
		return "T"
	}
	return "B"
}

//Minmax Scale
func (mm *minmax) Scale(val float64, start int) float64 {
	_, max := mm.SearchMax(start)
	_, min := mm.SearchMin(start)
	return (val - min) / (max - min)
}

//Inverts the min-max scaled value
func (mm *minmax) Invert(val float64, start int) float64 {
	_, max := mm.SearchMax(start)
	_, min := mm.SearchMin(start)
	return val*(max-min) + min
}

//高値（底値）から何パーセント折り返しているか返す。FIBO比較するために使う想定。
func (mm *minmax) Wrap(val float64, start int) float64 {
	maxi, mini := mm.SearchMinMax(start)
	scaled := mm.Scale(val, start)
	if maxi < mini {
		return scaled
	}
	return 1 - scaled
}

func (i *Inf) AddWrap(val float64) *Inf {
	i.Scaled = Wrap(val, i.Maxv, i.Minv, i.Maxi, i.Mini)
	return i
}

//
//普通の関数バージョン
//
func Scale(val, max, min float64) float64 {
	return (val - min) / (max - min)
}

func Invert(val, max, min float64) float64 {
	return val*(max-min) + min
}

func Wrap(val, maxv, minv float64, maxi, mini int) float64 {
	//高値（底値）から何パーセント折り返しているか返す。FIBO比較するために使う想定。
	scaled := Scale(val, maxv, minv)
	if maxi < mini {
		return scaled
	}
	return 1 - scaled
}

//
//ヘルパー
//
