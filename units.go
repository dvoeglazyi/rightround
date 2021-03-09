package gorewind

// kilometersInAU количество киллометров в астрономической единице.
// Примечание: принято строгое соответствие с заданной константой,
// за исключением эфемерид EPM до версии EPM2015, в которых задано
// собственное значение АЕ.
const kilometersInAU = 149597870.7

// secondsInDay количество секунд в сутках.
const secondsInDay = 24 * 60 * 60

// julianDate2000 12:00 1 января 2000 года в юлианских днях.
const julianDate2000 = 2451545

const maxPolynomialDegree = 20
