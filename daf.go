package gorewind

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"strings"
)

const (
	FormatSPK = iota
	FormatPCK
)

// DAFSegment сегмент файла в формате DAF.
type DAFSegment struct {
	offset      int32     // смещение сегмента
	length      int32     // длина сегмента
	dParameters []float64 // параметры double (float)
	iParameters []int32   // параметры int
	file        *os.File
}

const (
	representationPositionOnly = 2
	representationVelocityOnly = 20
)

type DAF struct {
	fileType          int
	dParametersNumber int // количество параметров double (float) в одном сегменте
	iParametersNumber int // количество параметров int в одном сегменте
	segments          []DAFSegment
	dParameters       []float64 // параметры double (float)
	iParameters       []int32   // параметры int
	file              *os.File
	buffer8           []byte
	buffer4           []byte
	name              string
}

func (d *DAF) readFloat() (float64, error) {
	if n, err := d.file.Read(d.buffer8); err != nil {
		return 0, err
	} else if n < len(d.buffer8) {
		return 0, errors.New("unexpected eof")
	}
	float := math.Float64frombits(binary.LittleEndian.Uint64(d.buffer8))
	return float, nil
}

func (d *DAF) readFloatToInt() (int, error) {
	if n, err := d.file.Read(d.buffer8); err != nil {
		return 0, err
	} else if n < len(d.buffer8) {
		return 0, errors.New("unexpected eof")
	}
	float := math.Float64frombits(binary.LittleEndian.Uint64(d.buffer8))
	if float-float64(int(float)) > 0 {
		return 0, errors.New("is not integer")
	}
	return int(float), nil
}

func (d *DAF) readInt32() (int32, error) {
	if n, err := d.file.Read(d.buffer4); err != nil {
		return 0, err
	} else if n < len(d.buffer4) {
		return 0, errors.New("unexpected eof")
	}
	return int32(binary.LittleEndian.Uint32(d.buffer4)), nil
}
func (d *DAF) readString(n int) (string, error) {
	b := make([]byte, n)
	if n, err := d.file.Read(b); err != nil {
		return "", err
	} else if n < len(b) {
		return "", errors.New("unexpected eof")
	}
	return string(b), nil
}

func newDAF(file *os.File) (*DAF, error) {
	d := DAF{
		file:    file,
		buffer4: make([]byte, 4),
		buffer8: make([]byte, 8),
	}

	id, err := d.readString(8)
	if err != nil {
		return nil, err
	}
	if strings.Contains(id, "DAF/SPK") || strings.Contains(id, "NAIF/DAF") {
		d.fileType = FormatSPK
	} else if strings.Contains(id, "DAF/PCK") {
		d.fileType = FormatPCK
	} else {
		return nil, errors.New("unsupported format")
	}
	return &d, nil
}

func (d *DAF) read() error {
	dParametersNumber, err := d.readInt32()
	if err != nil {
		return err
	}

	intParametersNumber, err := d.readInt32()
	if err != nil {
		return err
	} else if intParametersNumber < 2 {
		return errors.New("wrong format: i-parameter < 2")
	}
	d.dParametersNumber = int(dParametersNumber)
	d.iParametersNumber = int(intParametersNumber)
	usedIntParametersNumber := d.iParametersNumber - 2

	d.name, err = d.readString(60)
	if err != nil {
		return err
	}
	firstSummary, err := d.readInt32()
	if err != nil {
		return err
	}
	lastSummary, err := d.readInt32()
	if err != nil {
		return err
	}
	// check n < offset
	if _, err := d.file.Seek(4, io.SeekCurrent); err != nil {
		return err
	}
	// skip reserved blocks
	if _, err := d.file.Seek((int64(firstSummary)-2)*1024, io.SeekStart); err != nil {
		return err
	}

	summary := int(firstSummary)
	prevSummary := 0
	nSegments := 0
	for summary != 0 {
		if _, err := d.file.Seek((int64(summary)-1)*1024, io.SeekStart); err != nil {
			return err
		}
		nextSummary, err := d.readFloatToInt()
		if err != nil {
			return err
		}
		prev, err := d.readFloatToInt()
		if err != nil {
			return err
		} else if prev != prevSummary {
			return errors.New("bad format: prev summary is wrong")
		}
		nSummaries, err := d.readFloatToInt()
		if err != nil {
			return err
		}
		nSegments += nSummaries
		prevSummary = summary
		summary = nextSummary
	}
	if prevSummary != int(lastSummary) {
		return errors.New("previous summary is not equal last")
	}

	summary = int(firstSummary)
	prevSummary = 0
	for summary != 0 {
		if _, err := d.file.Seek((int64(summary)-1)*1024, io.SeekStart); err != nil {
			return err
		}
		nextSummary, err := d.readFloatToInt()
		if err != nil {
			return err
		}
		if _, err := d.file.Seek(8, io.SeekCurrent); err != nil {
			return err
		}
		nSummaries, err := d.readFloatToInt()
		if err != nil {
			return err
		}
		for num := nSummaries; num > 0; num-- {
			segment := DAFSegment{
				file:        d.file,
				dParameters: make([]float64, d.dParametersNumber),
				iParameters: make([]int32, usedIntParametersNumber),
			}

			for i := 0; i < d.dParametersNumber; i++ {
				if segment.dParameters[i], err = d.readFloat(); err != nil {
					return err
				}
			}
			for i := 0; i < usedIntParametersNumber; i++ {
				if segment.iParameters[i], err = d.readInt32(); err != nil {
					return err
				}
			}

			initialAddress, err := d.readInt32()
			if err != nil {
				return err
			}
			finalAddress, err := d.readInt32()
			if err != nil {
				return err
			}

			if d.dParametersNumber%2 != 0 {
				if _, err := d.file.Seek(4, io.SeekCurrent); err != nil {
					return err
				}
			}

			segment.offset = initialAddress - 1
			segment.length = finalAddress - initialAddress + 1
			d.segments = append(d.segments, segment)
		}
		prevSummary = summary
		summary = nextSummary
	}
	return nil
}

func (s *DAFSegment) readRange(start, length int) ([]float64, error) {
	const byteLength = 8
	if start+length > int(s.length) {
		return nil, errors.New("segment is out of file range")
	}
	if _, err := s.file.Seek((int64(s.offset)+int64(start))*byteLength, io.SeekStart); err != nil {
		return nil, err
	}
	buffer := make([]byte, length*byteLength)
	if n, err := s.file.Read(buffer); err != nil {
		return nil, err
	} else if n < length {
		return nil, io.EOF
	}
	result := make([]float64, length)
	for i := range result {
		result[i] = math.Float64frombits(binary.LittleEndian.Uint64(buffer[i*byteLength : i*byteLength+byteLength]))
	}
	return result, nil
}
