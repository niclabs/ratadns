package main

func Ema(number int, data []float64) []float64 {
	var sum float64 = 0
	var ema []float64

	//Calculate first estimation with SMA (average)
	for i := 0; i < number; i++ {
		sum += data[i]
	}
	sma := sum / float64(number)

	first := Emastep(float64(number), data[0], sma)
	ema = append(ema, first)

	//begin iteration
	for i := 1; i < len(data); i++ {
		value := Emastep(float64(number), data[i], ema[i-1])
		ema = append(ema, value)
	}

	return ema
}

func Emastep(number, actual, past float64) float64 {
	k := 2 / (number + 1)
	val := k*(actual-past) + past
	return val
}
