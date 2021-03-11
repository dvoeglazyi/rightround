package rightround

// combineTwoEphemeris выполняет линейную комбинацию двух эфемерид.
func (e *Ephemeris) combineTwoEphemeris(object1, basis1 int, object2, basis2 int, coefficient1 float64, coefficient2 float64, date0, date1 float64, withVelocity bool) (Coords, Coords, error) {
	coords1, velocity1, err := e.CalculateRectangularCoords(object1, basis1, date0, date1, withVelocity)
	if err != nil {
		return coords1, velocity1, err
	}
	coords2, velocity2, err := e.CalculateRectangularCoords(object2, basis2, date0, date1, withVelocity)
	if err != nil {
		return coords1, velocity2, err
	}

	coords := Coords{
		X: coords1.X*coefficient1 + coords2.X*coefficient2,
		Y: coords1.Y*coefficient1 + coords2.Y*coefficient2,
		Z: coords1.Z*coefficient1 + coords2.Z*coefficient2,
	}

	if !withVelocity {
		return coords, Coords{}, nil
	}

	return coords, Coords{
		X: velocity1.X*coefficient1 + velocity2.X*coefficient2,
		Y: velocity1.Y*coefficient1 + velocity2.Y*coefficient2,
		Z: velocity1.Z*coefficient1 + velocity2.Z*coefficient2,
	}, nil
}

// combineTwoEphemeris выполняет линейную комбинацию трёх эфемерид.
func (e *Ephemeris) combineThreeEphemeris(object1, basis1, object2, basis2, object3, basis3 int, coefficient1, coefficient2, coefficient3 float64, date0, date1 float64, withVelocity bool) (Coords, Coords, error) {
	coords1, velocity1, err := e.CalculateRectangularCoords(object1, basis1, date0, date1, withVelocity)
	if err != nil {
		return coords1, velocity1, err
	}
	coords2, velocity2, err := e.CalculateRectangularCoords(object2, basis2, date0, date1, withVelocity)
	if err != nil {
		return coords1, velocity2, err
	}
	coords3, velocity3, err := e.CalculateRectangularCoords(object3, basis3, date0, date1, withVelocity)
	if err != nil {
		return coords1, velocity2, err
	}

	coords := Coords{
		X: coords1.X*coefficient1 + coords2.X*coefficient2 + coords3.X*coefficient3,
		Y: coords1.Y*coefficient1 + coords2.Y*coefficient2 + coords3.Y*coefficient3,
		Z: coords1.Z*coefficient1 + coords2.Z*coefficient2 + coords3.Z*coefficient3,
	}

	if !withVelocity {
		return coords, Coords{}, nil
	}

	return coords, Coords{
		X: velocity1.X*coefficient1 + velocity2.X*coefficient2 + velocity3.X*coefficient3,
		Y: velocity1.Y*coefficient1 + velocity2.Y*coefficient2 + velocity3.Y*coefficient3,
		Z: velocity1.Z*coefficient1 + velocity2.Z*coefficient2 + velocity3.Z*coefficient3,
	}, nil
}
