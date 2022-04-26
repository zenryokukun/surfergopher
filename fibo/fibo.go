package fibo

func fibos() []float64 {
	return []float64{0.236, 0.382, 0.50, 0.618, 0.764}
}

func Level(scale float64) int {
	/*
		LEVEL0: scale < 0.236
		LEVEL1: 0.236 <= scale < 0.382
		LEVEL2: 0.382 <= scale <0.5
		LEVEL3: 0.5   <= scale < 0.618
		LEVEL4: 0.618 <= scale < 0.764
		LEVEL5: 0.764 <= scale
	*/
	fib := fibos()
	i := 0
	for {
		if i == len(fib) {
			return i
		}
		if scale < fib[i] {
			return i
		}
		i++
	}
}
