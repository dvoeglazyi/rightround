package gorewind

// Единицы измерения расстояния.
const (
	UnitCodeAU = 1 // астрономические единицы
	UnitCodeKM = 2 // киломметры
)

// Единицы измерения времени.
const (
	UnitCodeSec = 3 // световые секунды
	UnitCodeDay = 4 // световые дни
)

// Числовые коды небесных тел и других объектов.
// Нумерация соответствует принятой в формате SPK.
const (
	EphemerisSunSystem = 0   // барицентр Солнечной системы
	EphemerisMercury   = 1   // Меркурий
	EphemerisVenus     = 2   // Венера
	EphemerisEarthMoon = 3   // барицентр системы Земля-Луна
	EphemerisMars      = 4   // барицентр системы Марса
	EphemerisJupiter   = 5   // барицентр системы Юпитера
	EphemerisSaturn    = 6   // барицентр системы Сатурна
	EphemerisUranus    = 7   // барицентр системы Урана
	EphemerisNeptune   = 8   // барицентр системы Нептуна
	EphemerisPluto     = 9   // барицентр системы Плутона
	EphemerisSun       = 10  // Солнце
	EphemerisMoon      = 301 // Луна
	EphemerisEarth     = 399 // Земля
)

// Разность шкал TT - TDB.
const EphemerisCodeMinusTDB = 1000000001

// Числовые коды лунных систем координат, принятых в различных эфемеридах.
// Нумерация соответствует соглашениям, принятым производителями эфемерид в формате PCK.
const (
	EphemerisMoonPrincipalAxesDE403   = 31002
	EphemerisMoonPrincipalAxesDE421   = 31006
	EphemerisMoonPrincipalAxesDE430   = 32006
	EphemerisMoonPrincipalAxesInPOP   = 1900301
	EphemerisMoonPrincipalAxesEPM2011 = 1800301
	EphemerisMoonPrincipalAxesEPM2015 = 1800302
	EphemerisMoonPrincipalAxesEPM2017 = 1800303
)
