### rightround
Библиотека на Golang для работы с эфемеридами в форматах семейства SPICE: [EPM](http://iaaras.ru/dept/ephemeris/epm/) и [DE](https://ssd.jpl.nasa.gov/?ephemerides).

#### Пример использования
```
const (
    pathEPM = "ephemeris/epm2017h.bsp"
    pathDE = "ephemeris/de441.bsp"
)

// загрузка эфемерид EPM2017
ephemeris := rightround.NewEphemeris()
if err := ephemeris.LoadFile(pathEPM); err != nil {
    return err
}

julianDays, julianTime := 2459395.0, 0.5
// рассчёт положения Меркурия относительно Земли
// на момент 00:00:00 30 июня 2021 г.
coords, _, err := ephemeris.CalculateRectangularCoordsAndScaleVelocity(gorewind.EphemerisMercury, gorewind.EphemerisEarth, julianDays, julianTime, false)
if err != nil {
    return err
}
// вывод в километрах
fmt.Printf("%.5f %.5f %.5f\n", coords.X, coords.Y, coords.Z)
// результат: -151786440.78263 -28597178.81489 -18024058.24283
```

#### Список источников
* [Библиотека ephemeris-access](https://gitlab.iaaras.ru/iaaras/ephemeris-access) (на языке C) / Дмитрий Павлов / ИПА РАН