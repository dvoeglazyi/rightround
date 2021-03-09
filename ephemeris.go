package gorewind

import (
	"fmt"
	"os"
)

type Ephemeris struct {
	dafs                  []*DAF
	theories              []*Theory
	distanceScalingFactor float64
	timeScalingFactor     float64
	distanceUnits         int

	allocatedTheoriesCount int
	// флаги, описывающие представление системы Земля-Луна в загруженных эфемеридах
	haveEarthRefSunSystem     bool
	haveMoonRefEarth          bool
	haveMoonRefEarthMoon      bool
	haveEarthRefEarthMoon     bool
	haveEarthMoonRefSunSystem bool

	leftmostJulianDate  float64
	rightmostJulianDate float64
}

func NewEphemeris() *Ephemeris {
	return &Ephemeris{
		distanceUnits:         UnitCodeKM,
		timeScalingFactor:     secondsInDay,
		distanceScalingFactor: 1,
		leftmostJulianDate:    -1,
		rightmostJulianDate:   -1,
	}
}

func (e *Ephemeris) LoadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	daf, err := newDAF(file)
	if err != nil {
		return err
	}

	return e.loadDAF(daf)
}

func (e *Ephemeris) loadDAF(daf *DAF) error {
	if err := daf.read(); err != nil {
		return err
	}
	for i, segment := range daf.segments {
		var theory Theory
		if daf.fileType == FormatSPK {
			theory.object = int(segment.iParameters[0])
			theory.basis = int(segment.iParameters[1])
			theory.representation = int(segment.iParameters[3])

			e.haveMoonRefEarth = e.haveMoonRefEarth || (theory.object == EphemerisMoon && theory.basis == EphemerisEarth)
			e.haveMoonRefEarthMoon = e.haveMoonRefEarthMoon || (theory.object == EphemerisMoon && theory.basis == EphemerisEarthMoon)
			e.haveEarthRefEarthMoon = e.haveEarthRefEarthMoon || (theory.object == EphemerisEarth && theory.basis == EphemerisEarthMoon)
			e.haveEarthRefSunSystem = e.haveEarthRefSunSystem || (theory.object == EphemerisEarth && theory.basis == EphemerisSunSystem)
			e.haveEarthMoonRefSunSystem = e.haveEarthMoonRefSunSystem || (theory.object == EphemerisEarthMoon && theory.basis == EphemerisSunSystem)
		} else if daf.fileType == FormatPCK {
			theory.object = int(segment.iParameters[0])
			theory.representation = int(segment.iParameters[2])
		}

		if theory.representation == representationPositionOnly {
			params, err := segment.readRange(int(segment.length)-4, 4)
			if err != nil {
				return err
			}
			// params[0] точка отсчёта в теории, в секундах начиная с J2000
			// конвертация секунд в целочисленные дни (JulianDays) + фракция от дней (JulianDaysMod)
			days := int(params[0] / secondsInDay)
			theory.julianDays = julianDate2000 + float64(days)
			theory.julianDaysMod = (params[0] - float64(days)*secondsInDay) / secondsInDay // check
			theory.intervalLen = float64(int(params[1]) / secondsInDay)                    // check why int
			theory.rSize = int(params[2])
			theory.nIntervals = int(params[3])

			// проверить, что RSize в 3N + 2
			if theory.rSize%3 != 2 {
				return fmt.Errorf("bad rSize (%d)", theory.rSize)
			}
			// полиномиальный градус в N-1
			theory.polynomialDegree = (theory.rSize-2)/3 - 1
			// dScale и tScale не используются в этом типе ефемерид
			theory.dScale = 1
			theory.tScale = 1
		} else if theory.representation == representationVelocityOnly {
			params, err := segment.readRange(int(segment.length)-7, 7)
			if err != nil {
				return err
			}

			theory.dScale = params[0]
			theory.tScale = params[1] / secondsInDay
			theory.julianDays = params[2]
			theory.julianDaysMod = params[3]
			theory.intervalLen = float64(int(params[4]))
			theory.rSize = int(params[5])
			theory.nIntervals = int(params[6])
			// rSize в 3N
			if theory.rSize%3 != 0 {
				return fmt.Errorf("bad RSize (%d)", theory.rSize)
			}
			// полиномиальный градус в N-2
			theory.polynomialDegree = theory.rSize/3 - 2
		} else {
			return fmt.Errorf("unsupported representation (%d)", theory.representation)
		}

		if theory.polynomialDegree > maxPolynomialDegree {
			return fmt.Errorf("polynomial degree limit (%d) exceeded in file", theory.polynomialDegree)
		}
		theory.cachedInterval = -1
		theory.fileType = daf.fileType
		theory.segment = &daf.segments[i]

		if leftmostDate := theory.julianDays + theory.julianDaysMod; e.leftmostJulianDate < 0 || e.leftmostJulianDate > leftmostDate {
			e.leftmostJulianDate = leftmostDate
		}
		if rightmostDate := theory.julianDays + theory.julianDaysMod + theory.intervalLen*float64(theory.nIntervals); e.rightmostJulianDate < 0 || e.rightmostJulianDate < rightmostDate {
			e.rightmostJulianDate = rightmostDate
		}
		e.theories = append(e.theories, &theory)
	}
	e.dafs = append(e.dafs, daf)
	return nil
}

// setDistanceUnits устанавливает единицы измерения расстояния.
func (e *Ephemeris) setDistanceUnits(unit int) error {
	if unit == UnitCodeKM {
		// в SPK файлах уже в киллометрах
		e.distanceScalingFactor = 1
	} else if unit == UnitCodeAU {
		e.distanceScalingFactor = 1 / kilometersInAU
	} else {
		return fmt.Errorf("unknown distance units: %d", unit)
	}
	e.distanceUnits = unit
	return nil
}

// setTimeUnits устанавливает единицы измерения времени.
func (e *Ephemeris) setTimeUnits(unit int) error {
	if unit == UnitCodeDay {
		// внутренние единицы измерения SPK/PCK это дни
		e.timeScalingFactor = 1
	} else if unit == UnitCodeSec { // в SPK бывают секунды
		e.timeScalingFactor = secondsInDay
	} else {
		return fmt.Errorf("unknown time units: %d", unit)
	}
	return nil
}
