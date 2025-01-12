package decor

import (
	"fmt"
	"io"
	"strconv"

	"github.com/mikewiacek/mpb/internal"
)

type percentageType float64

func (s percentageType) Format(st fmt.State, verb rune) {
	var prec int
	switch verb {
	case 'd':
	case 's':
		prec = -1
	default:
		if p, ok := st.Precision(); ok {
			prec = p
		} else {
			prec = 6
		}
	}

	io.WriteString(st, strconv.FormatFloat(float64(s), 'f', prec, 64))

	if st.Flag(' ') {
		io.WriteString(st, " ")
	}
	io.WriteString(st, "%")
}

// Percentage returns percentage decorator. It's a wrapper of NewPercentage.
func Percentage(wcc ...WC) Decorator {
	return NewPercentage("% d", wcc...)
}

// NewPercentage percentage decorator with custom fmt string.
//
// fmt examples:
//
//	fmt="%.1f"  output: "1.0%"
//	fmt="% .1f" output: "1.0 %"
//	fmt="%d"    output: "1%"
//	fmt="% d"   output: "1 %"
//
func NewPercentage(fmt string, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	if fmt == "" {
		fmt = "% d"
	}
	d := &percentageDecorator{
		WC:  wc,
		fmt: fmt,
	}
	return d
}

type percentageDecorator struct {
	WC
	fmt         string
	completeMsg *string
}

func (d *percentageDecorator) Decor(st *Statistics) string {
	if st.Completed && d.completeMsg != nil {
		return d.FormatMsg(*d.completeMsg)
	}
	p := internal.Percentage(st.Total, st.Current, 100)
	return d.FormatMsg(fmt.Sprintf(d.fmt, percentageType(p)))
}

func (d *percentageDecorator) OnCompleteMessage(msg string) {
	d.completeMsg = &msg
}
