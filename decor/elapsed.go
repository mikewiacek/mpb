package decor

import (
	"time"
)

// Elapsed decorator. It's wrapper of NewElapsed.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`wcc` optional WC config
func Elapsed(style TimeStyle, wcc ...WC) Decorator {
	return NewElapsed(style, time.Now(), wcc...)
}

// NewElapsed returns elapsed time decorator.
//
//	`style` one of [ET_STYLE_GO|ET_STYLE_HHMMSS|ET_STYLE_HHMM|ET_STYLE_MMSS]
//
//	`startTime` start time
//
//	`wcc` optional WC config
func NewElapsed(style TimeStyle, startTime time.Time, wcc ...WC) Decorator {
	var wc WC
	for _, widthConf := range wcc {
		wc = widthConf
	}
	wc.Init()
	d := &elapsedDecorator{
		WC:        wc,
		startTime: startTime,
		producer:  chooseTimeProducer(style),
	}
	return d
}

type elapsedDecorator struct {
	WC
	startTime   time.Time
	producer    func(time.Duration) string
	msg         string
	completeMsg *string
}

func (d *elapsedDecorator) Decor(st *Statistics) string {
	if st.Completed {
		if d.completeMsg != nil {
			return d.FormatMsg(*d.completeMsg)
		}
		return d.FormatMsg(d.msg)
	}

	d.msg = d.producer(time.Since(d.startTime))
	return d.FormatMsg(d.msg)
}

func (d *elapsedDecorator) OnCompleteMessage(msg string) {
	d.completeMsg = &msg
}
