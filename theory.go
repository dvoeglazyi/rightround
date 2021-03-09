package gorewind

import "math"

type Theory struct {
	segment            *DAFSegment
	object             int
	basis              int
	representation     int
	julianDays         float64
	julianDaysMod      float64
	dScale             float64
	tScale             float64
	intervalLen        float64
	rSize              int
	nIntervals         int
	polynomialDegree   int
	cachedInterval     int
	cachedCoefficients []float64
	fileType           int
}

// findInterval возвращает номер интервала, которому принадлежит юлианская дата, и число от -1 до 1, которое описывает позицию внутри интервала.
func (t *Theory) findInterval(date1, date2 float64) (int, float64) {
	diff := date1 + date2 - t.julianDays - t.julianDaysMod
	interval := int(math.Floor(diff / t.intervalLen))
	diffInterval := diff - float64(interval)*t.intervalLen
	return interval, (diffInterval/t.intervalLen)*2 - 1
}

// isDateInRange проверяет, входит ли дата в диапазон, который охватывает теория.
func (t *Theory) isDateInRange(date1, date2 float64) bool {
	diff := date1 + date2 - t.julianDays - t.julianDaysMod
	return diff >= 0 && diff <= float64(t.nIntervals)*t.intervalLen
}
