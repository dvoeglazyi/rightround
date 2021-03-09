package gorewind

import (
	"errors"
	"fmt"
	"math"
)

// CalculateRectangularCoordsAndScaleVelocity вычисляет прямоугольные координаты и скорости на заданную дату.
func (e *Ephemeris) CalculateRectangularCoordsAndScaleVelocity(object, basis int, date1, date2 float64, withVelocity bool) (Coords, Coords, error) {
	coords, velocity, err := e.CalculateRectangularCoords(object, basis, date1, date2, withVelocity)
	if err != nil {
		return coords, velocity, err
	}
	// масштабирование дистанции (координат) уже выполнено в calculateByTheory
	if withVelocity {
		velocity.X /= e.timeScalingFactor
		velocity.Y /= e.timeScalingFactor
		velocity.Z /= e.timeScalingFactor
	}
	return coords, velocity, nil
}

// CalculateEulerAngles вычисляет эйлеровы углы и скорости их изменения на заданную дату.
func (e *Ephemeris) CalculateEulerAngles(frame int, date1, date2 float64, withRates bool) (Coords, Coords, error) {
	// поиск нужного сегмента
	var theory, singleTheory *Theory
	isSingle := true
	for _, t := range e.theories {
		// проверка, что заданная дата - дата начала фрейма в диапазоне [0, длина сегмента]
		if !t.isDateInRange(date1, date2) {
			continue
		}
		if t.object == frame {
			theory = t
			break
		}
		// проверка, что это одиночная PCK-теория
		if frame == 0 && t.fileType == FormatPCK {
			if singleTheory == nil {
				singleTheory = t
			} else if isSingle && singleTheory.object != t.object {
				isSingle = false
			}
		}
	}

	if theory == nil {
		if singleTheory == nil {
			return Coords{}, Coords{}, fmt.Errorf("theory for frame %d not found", frame)
		}
		theory = singleTheory
	}

	angles, rates, err := e.calculateByTheory(theory, date1, date2, false, withRates)
	if err != nil {
		return Coords{}, Coords{}, err
	}

	if withRates {
		rates.X /= e.timeScalingFactor
		rates.Y /= e.timeScalingFactor
		rates.Z /= e.timeScalingFactor
	}

	return angles, rates, nil
}

// CalculateTimeDiff вычисляет разности шкал времени на заданную дату.
func (e *Ephemeris) CalculateTimeDiff(code int, date1, date2 float64) (float64, error) {
	var theory *Theory
	for _, t := range e.theories {
		if t.object == code && t.isDateInRange(date1, date2) {
			theory = t
			break
		}
	}
	if theory == nil {
		return 0, fmt.Errorf("theory for time difference %d not found", code)
	}
	coords, _, err := e.calculateByTheory(theory, date1, date2, false, false)
	if err != nil {
		return 0, err
	}
	// применение фактора масштабирования к временной разнице
	// (TT-TDB хранится в секундах, поэтому выполняется деление на кол-во секунд в днях, чтобы получить дни
	// а затем применить фактор масштабирования)
	diff := coords.X * e.timeScalingFactor / secondsInDay
	return diff, err
}

// CalculateRectangularCoords вычисляет прямоугольные координаты для заданного объекта относительно заданного объекта и выполняет масштабирование.
func (e *Ephemeris) CalculateRectangularCoords(object, basis int, date1, date2 float64, withVelocity bool) (Coords, Coords, error) {
	if object == basis {
		return Coords{}, Coords{}, nil
	}
	if object == EphemerisEarth && basis == EphemerisSunSystem {
		if !e.haveEarthRefSunSystem && e.haveEarthRefEarthMoon && e.haveEarthMoonRefSunSystem {
			return e.combineTwoEphemeris(EphemerisEarth, EphemerisEarthMoon, EphemerisEarthMoon, EphemerisSunSystem, 1, 1, date1, date2, withVelocity)
		}
	} else if object == EphemerisMoon && basis == EphemerisSunSystem {
		if e.haveMoonRefEarth && e.haveEarthRefSunSystem {
			return e.combineTwoEphemeris(EphemerisMoon, EphemerisEarth, EphemerisEarthMoon, EphemerisSunSystem, 1, 1, date1, date2, withVelocity)
		} else if e.haveMoonRefEarth && e.haveEarthMoonRefSunSystem {
			return e.combineTwoEphemeris(EphemerisMoon, EphemerisEarthMoon, EphemerisEarthMoon, EphemerisSunSystem, 1, 1, date1, date2, withVelocity)
		} else if e.haveMoonRefEarth && e.haveEarthRefEarthMoon && e.haveEarthMoonRefSunSystem {
			return e.combineThreeEphemeris(EphemerisMoon, EphemerisEarth, EphemerisEarth, EphemerisEarthMoon, EphemerisEarthMoon, EphemerisSunSystem, 1, 1, 1, date1, date2, withVelocity)
		}
	} else if object == EphemerisMoon && basis == EphemerisEarth {
		if !e.haveMoonRefEarth && e.haveMoonRefEarthMoon && e.haveEarthRefEarthMoon {
			return e.combineTwoEphemeris(EphemerisMoon, EphemerisEarthMoon, EphemerisEarth, EphemerisEarthMoon, 1, -1, date1, date2, withVelocity)
		}
	} else if (object == EphemerisEarth && basis == EphemerisEarthMoon) ||
		(object == EphemerisMoon && basis == EphemerisEarthMoon) ||
		(object >= EphemerisMercury && object <= EphemerisSun && basis == EphemerisSunSystem) {

	} else if (object == EphemerisSunSystem && basis >= EphemerisMercury && basis <= EphemerisSun) ||
		(object == EphemerisSunSystem && (basis == EphemerisMoon || basis == EphemerisEarth)) ||
		(object == EphemerisEarth && basis == EphemerisMoon) {
		// поменять местами объект и базу
		coords, velocity, err := e.CalculateRectangularCoords(basis, object, date1, date2, withVelocity)
		if err != nil {
			return Coords{}, Coords{}, err
		}
		coords.X = -coords.X
		coords.Y = -coords.Y
		coords.Z = -coords.Z
		if withVelocity {
			velocity.X = -velocity.X
			velocity.Y = -velocity.Y
			velocity.Z = -velocity.Z
		}
		return coords, velocity, nil

	} else if object != EphemerisSunSystem && basis != EphemerisSunSystem {
		coords, velocity, err := e.CalculateRectangularCoords(object, EphemerisSunSystem, date1, date2, withVelocity)
		if err != nil {
			return Coords{}, Coords{}, err
		}
		basisCoords, basisVelocity, err := e.CalculateRectangularCoords(basis, EphemerisSunSystem, date1, date2, withVelocity)
		if err != nil {
			return Coords{}, Coords{}, err
		}
		coords.X -= basisCoords.X
		coords.Y -= basisCoords.Y
		coords.Z -= basisCoords.Z
		if withVelocity {
			velocity.X -= basisVelocity.X
			velocity.Y -= basisVelocity.Y
			velocity.Z -= basisVelocity.Z
		}
		return coords, velocity, nil
	}
	var theory *Theory
	for _, t := range e.theories {
		if t.object == object && t.basis == basis && t.isDateInRange(date1, date2) {
			theory = t
			break
		}
	}
	if theory == nil {
		return Coords{}, Coords{}, errors.New("theory for object and reference not found")
	}

	return e.calculateByTheory(theory, date1, date2, true, withVelocity)
}

// calculateByTheory вычисляет прямоугольные координаты для заданной теории и даты.
func (e *Ephemeris) calculateByTheory(theory *Theory, date1, date2 float64, scaleDistance, withVelocity bool) (Coords, Coords, error) {
	interval, posInInterval := theory.findInterval(date1, date2)

	coefficients := theory.cachedCoefficients
	toRead := theory.cachedInterval != interval

	var coords, velocity Coords
	if theory.representation == representationPositionOnly {
		polynomials := calcChebyshevPolynomials(theory.polynomialDegree+1, posInInterval)

		if toRead {
			var err error
			if coefficients, err = theory.segment.readRange(theory.rSize*interval+2, theory.rSize-2); err != nil {
				return coords, velocity, err
			}
		}

		for i := theory.polynomialDegree; i >= 0; i-- {
			coords.X += polynomials[i] * coefficients[i]
			coords.Y += polynomials[i] * coefficients[i+theory.polynomialDegree+1]
			coords.Z += polynomials[i] * coefficients[i+(theory.polynomialDegree+1)*2]
		}
		if scaleDistance {
			coords.X *= e.distanceScalingFactor
			coords.Y *= e.distanceScalingFactor
			coords.Z *= e.distanceScalingFactor
		}
		if withVelocity {
			for i := theory.polynomialDegree; i >= 0; i-- {
				velocity.X += polynomials[i] * coefficients[i]
				velocity.Y += polynomials[i] * coefficients[i+theory.polynomialDegree+1]
				velocity.Z += polynomials[i] * coefficients[i+(theory.polynomialDegree+1)*2]
			}
			velocity.X /= 0.5 * theory.intervalLen
			velocity.Y /= 0.5 * theory.intervalLen
			velocity.Z /= 0.5 * theory.intervalLen

			if scaleDistance {
				velocity.X *= e.distanceScalingFactor
				velocity.Y *= e.distanceScalingFactor
				velocity.Z *= e.distanceScalingFactor
			}
		}
	} else if theory.representation == representationVelocityOnly {
		distanceScale := theory.dScale * e.distanceScalingFactor

		if scaleDistance && e.distanceUnits == UnitCodeAU && math.Abs(theory.dScale-kilometersInAU) < 1000 {
			// теория имеет собственное значение астрономических единиц, оставить как есть
			distanceScale = 1
		}

		polynomials := calcChebyshevPolynomials(theory.polynomialDegree+2, posInInterval)
		antiDerivatives := calcChebyshevAntiDerivatives(theory.polynomialDegree+1, posInInterval, polynomials)

		if toRead {
			var err error
			if coefficients, err = theory.segment.readRange(theory.rSize*interval, theory.rSize); err != nil {
				return coords, velocity, err
			}
		}

		for i := theory.polynomialDegree; i >= 0; i-- {
			coords.X += antiDerivatives[i] * coefficients[i]
			coords.Y += antiDerivatives[i] * coefficients[i+theory.polynomialDegree+2]
			coords.Z += antiDerivatives[i] * coefficients[i+(theory.polynomialDegree+2)*2]
		}
		coords.X = 0.5*theory.intervalLen*coords.X + coefficients[theory.polynomialDegree+1]
		coords.Y = 0.5*theory.intervalLen*coords.Y + coefficients[theory.polynomialDegree+1+theory.polynomialDegree+2]
		coords.Z = 0.5*theory.intervalLen*coords.Z + coefficients[theory.polynomialDegree+1+(theory.polynomialDegree+2)*2]

		if scaleDistance {
			coords.X *= distanceScale
			coords.Y *= distanceScale
			coords.Z *= distanceScale
		}

		if withVelocity {
			for i := theory.polynomialDegree; i >= 0; i-- {
				velocity.X += polynomials[i] * coefficients[i]
				velocity.Y += polynomials[i] * coefficients[i+theory.polynomialDegree+2]
				velocity.Z += polynomials[i] * coefficients[i+(theory.polynomialDegree+2)*2]
			}

			if scaleDistance {
				velocity.X *= distanceScale / theory.tScale
				velocity.Y *= distanceScale / theory.tScale
				velocity.Z *= distanceScale / theory.tScale
			}
		}
	}
	theory.cachedInterval = interval
	theory.cachedCoefficients = coefficients
	return coords, velocity, nil
}
