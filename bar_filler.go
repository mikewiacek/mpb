package mpb

import (
	"io"
	"unicode/utf8"

	"github.com/mikewiacek/mpb/decor"
	"github.com/mikewiacek/mpb/internal"
)

const (
	rLeft = iota
	rFill
	rTip
	rEmpty
	rRight
	rRevTip
	rRefill
)

var defaultBarStyle = "[=>-]<+"

type barFiller struct {
	format       [][]byte
	refillAmount int64
	reverse      bool
	noBrackets   bool
}

func newDefaultBarFiller() Filler {
	bf := &barFiller{
		format: make([][]byte, utf8.RuneCountInString(defaultBarStyle)),
	}
	bf.setStyle(defaultBarStyle)
	return bf
}

func (s *barFiller) setStyle(style string) {
	if !utf8.ValidString(style) {
		return
	}
	src := make([][]byte, 0, utf8.RuneCountInString(style))
	for _, r := range style {
		src = append(src, []byte(string(r)))
	}
	copy(s.format, src)
}

func (s *barFiller) SetRefill(amount int64) {
	s.refillAmount = amount
}

func (s *barFiller) Fill(w io.Writer, width int, stat *decor.Statistics) {

	if !s.noBrackets {
		// don't count rLeft and rRight as progress
		width -= 2
		if width < 2 {
			return
		}
		w.Write(s.format[rLeft])
		defer w.Write(s.format[rRight])
	}

	bb := make([][]byte, width)

	cwidth := int(internal.PercentageRound(stat.Total, stat.Current, width))

	for i := 0; i < cwidth; i++ {
		bb[i] = s.format[rFill]
	}

	if s.refillAmount > 0 {
		var rwidth int
		if s.refillAmount > stat.Current {
			rwidth = cwidth
		} else {
			rwidth = int(internal.PercentageRound(stat.Total, int64(s.refillAmount), width))
		}
		for i := 0; i < rwidth; i++ {
			bb[i] = s.format[rRefill]
		}
	}

	if cwidth > 0 && cwidth < width {
		bb[cwidth-1] = s.format[rTip]
	}

	for i := cwidth; i < width; i++ {
		bb[i] = s.format[rEmpty]
	}

	if s.reverse {
		if cwidth > 0 && cwidth < width {
			bb[cwidth-1] = s.format[rRevTip]
		}
		for i := len(bb) - 1; i >= 0; i-- {
			w.Write(bb[i])
		}
	} else {
		for i := 0; i < len(bb); i++ {
			w.Write(bb[i])
		}
	}
}
