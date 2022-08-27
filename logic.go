package main

import (
	"fmt"

	"github.com/zenryokukun/surfergopher/fibo"
	"github.com/zenryokukun/surfergopher/minmax"
)

//現在値が最大値を超えていれば"BUY"
//最小値を下回っていれば"SELL"を返す
//v:現在値、inf:最大値、最小値を保持しているstruct
func breakThrough(v float64, inf *minmax.Inf) string {
	if v > inf.Maxv {
		return "BUY"
	}
	if v < inf.Minv {
		return "SELL"
	}
	return ""
}

//fiboレベルに応じた新規取引判定
func fib(inf *minmax.Inf) string {
	lvl := fibo.Level(inf.Scaled)
	dec := ""
	if lvl == 0 {
		if inf.Which == "B" {
			dec = "BUY"
		} else if inf.Which == "T" {
			dec = "SELL"
		}
		logger(fmt.Sprintf("fibLvl:0. recent:%v scaled:%v decision:%v", inf.Which, inf.Scaled, dec))
	}

	if lvl == 2 {
		if inf.Which == "B" {
			dec = "SELL"
		} else if inf.Which == "T" {
			dec = "BUY"
		}
		logger(fmt.Sprintf("fibLvl:2. recent:%v scaled:%v decision:%v", inf.Which, inf.Scaled, dec))
	}

	if lvl == 3 {
		if inf.Which == "B" {
			dec = "SELL"
		} else if inf.Which == "T" {
			dec = "BUY"
		}
		logger(fmt.Sprintf("fibLvl:3. recent:%v scaled:%v decision:%v", inf.Which, inf.Scaled, dec))
	}

	if lvl == 4 {
		if inf.Which == "B" {
			dec = "SELL"
		} else if inf.Which == "T" {
			dec = "BUY"
		}
		logger(fmt.Sprintf("fibLvl:4. recent:%v scaled:%v decision:%v", inf.Which, inf.Scaled, dec))
	}

	return dec
}
