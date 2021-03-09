package gorewind

// calcChebyshevPolynomials вычисляет значения полинома Чебышева.
// positionInInterval - позиция внутри интервала [-1..1].
func calcChebyshevPolynomials(nCoefficients int, positionInInterval float64) []float64 {
	result := make([]float64, nCoefficients)
	result[0] = 1
	result[1] = positionInInterval
	for i := 2; i < nCoefficients; i++ {
		result[i] = 2*result[i-1]*positionInInterval - result[i-2]
	}
	return result
}

// calcChebyshevAntiDerivatives вычисляет значения первообразных полинома Чебышева.
// positionInInterval - позиция внутри интервала [-1..1].
func calcChebyshevAntiDerivatives(nCoefficients int, positionInInterval float64, polynomials []float64) []float64 {
	result := make([]float64, nCoefficients)
	result[0] = positionInInterval
	result[1] = (polynomials[2] + polynomials[0]) * 0.25
	for i := 2; i < nCoefficients; i++ {
		result[i] = 0.5 * (polynomials[i+1]/float64(i+1) - polynomials[i-1]/float64(i-1))
	}
	flag := false
	d := 2 * positionInInterval
	for i, j := 3, 1; i < nCoefficients; i, j = i+2, j+1 {
		d = 0.25/float64(j) + 0.25/float64(j+1)
		if flag = !flag; flag {
			d = -d
		}
		result[i] += d
	}
	return result
}

// calcChebyshevDerivatives вычисляет значения производных полинома Чебышева.
// positionInInterval - позиция внутри интервала [-1..1].
func calcChebyshevDerivatives(nCoefficients int, positionInInterval float64, polynomials []float64) []float64 {
	result := make([]float64, nCoefficients)
	result[0] = 0
	result[1] = 1
	for i := 2; i < nCoefficients; i++ {
		result[i] = positionInInterval*2*result[i-1] + 2*polynomials[i-1] - result[i-2]
	}
	return result
}
