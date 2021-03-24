package main

func Ema(number int, data []float64) []float64 {
	var k float64 = 2 / (float64(number) + 1)
	var sum float64 = 0
	var ema []float64

	//Calculate first estimation with SMA (average)
	for i := 0; i < number; i++ {
		sum += data[i]
	}
	sma := sum / float64(number)

	first := Emastep(k, data[0], sma)
	ema = append(ema, first)

	//begin iteration
	for i := 1; i < len(data); i++ {
		value := Emastep(k, data[i], ema[i-1])
		ema = append(ema, value)
	}

	return ema
}

func Emastep(k, actual, past float64) float64 {
	val := (k * (actual - past)) + past
	return val
}

func Errorfunc(data, esti float64) float64{
	abs := data - esti
	rel := (abs/data)*100
	return rel
}