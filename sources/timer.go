package relay

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

type Timeout interface {
	Timeout(Timer) error
}

type TimeoutFunc func(timer Timer) error

func (this TimeoutFunc) Timeout(timer Timer) error {
	return this(timer)
}

type Timer interface {
	Running() bool
	StartNow(interval TimerSchedule)
	Start(now time.Time, interval TimerSchedule)
	Stop()
}

type TimerSchedule interface {
	Next(time.Time) time.Time
}

func NewTimer(timeout Timeout) Timer {
	loop := InLoop()
	result := &timer{
		loop:    loop,
		running: 0,
		index:   -1,
	}
	result.executor = ExecFunc(func() error {
		return timeout.Timeout(result)
	})
	return result
}

func Schedule(now time.Time, cron string, action func() error) (func(), error) {
	schedule, err := WithCron(cron)
	if err != nil {
		return nil, err
	}
	timer := NewTimer(TimeoutFunc(func(timer Timer) error {
		return action()
	}))
	timer.Start(now, schedule)
	return func() {
		timer.Stop()
	}, nil
}

func After(d time.Duration, action func() error) func() {
	loop := InLoop()
	result := reusablelist.Get()
	if result == nil {
		result = reusableNew()
	}
	timer := result.(*reusable)
	timer.loop = loop
	timer.action = action
	timer.StartNow(WithDelay(d))
	return timer.stop
}

func WithDelay(delay time.Duration) TimerSchedule {
	return &linear{
		delay:    delay,
		interval: 0,
	}
}

func WithInterval(interval time.Duration) TimerSchedule {
	return &linear{
		delay:    interval,
		interval: interval,
	}
}

func WithDelayAndInterval(delay time.Duration, interval time.Duration) TimerSchedule {
	return &linear{
		delay:    delay,
		interval: interval,
	}
}

func WithCron(cron string) (TimerSchedule, error) {
	return cronparser.Parse(cron)
}

type timer struct {
	loop     Loop
	schedule TimerSchedule
	next     time.Time
	running  int32
	index    int
	executor Executor
}

func (this *timer) Running() bool {
	return atomic.LoadInt32(&this.running) != 0
}

func (this *timer) StartNow(schedule TimerSchedule) {
	this.Start(time.Now(), schedule)
}

func (this *timer) Start(now time.Time, schedule TimerSchedule) {
	if IsInLoop() == this.loop {
		timeronce.Do(func() {
			heap.Init(&timers)
			go func() {
				for {
					var now time.Time
					select {
					case modify := <-timers.modify:
						if modify.start {
							if modify.timer.index == -1 {
								heap.Push(&timers, modify.timer)
							}
						} else {
							if modify.timer.index != -1 {
								heap.Remove(&timers, modify.timer.index)
							}
						}
						now = time.Now()
					case now = <-timers.timer.C:
						for timers.length != 0 {
							timer := timers.timers[0]
							if timer.next.After(now) {
								break
							}
							heap.Pop(&timers)
							timer.next = timer.schedule.Next(now)
							if !timer.next.IsZero() {
								heap.Push(&timers, timer)
							} else {
								atomic.StoreInt32(&timer.running, 0)
							}
							timer.loop.Execute(timer.executor)
						}
					}
					interval := time.Hour * 24
					if len(timers.timers) != 0 {
						interval = timers.timers[0].next.Sub(now)
					}
					timers.timer.Reset(interval)
				}
			}()
		})
		this.schedule = schedule
		this.next = this.schedule.Next(now)
		if atomic.CompareAndSwapInt32(&this.running, 0, 1) {
			timers.modify <- timermodify{true, this}
		}
	} else {
		this.loop.Execute(ExecFunc(func() error {
			this.Start(now, schedule)
			return nil
		}))
	}
}

func (this *timer) Stop() {
	if IsInLoop() == this.loop {
		if atomic.CompareAndSwapInt32(&this.running, 1, 0) {
			timers.modify <- timermodify{false, this}
		}
	} else {
		this.loop.Execute(ExecFunc(func() error {
			this.Stop()
			return nil
		}))
	}
}

type reusable struct {
	timer
	action func() error
	stop   func()
}

type linear struct {
	last     *time.Time
	delay    time.Duration
	interval time.Duration
}

func (this *linear) Next(now time.Time) time.Time {
	if this.last == nil {
		result := now.Add(this.delay)
		this.last = &result
	} else if this.interval > 0 {
		result := this.last.Add(this.interval)
		this.last = &result
	} else {
		result := time.Time{}
		this.last = &result
	}
	return *this.last
}

type timermodify struct {
	start bool
	timer *timer
}

type timerlist struct {
	timers []*timer
	timer  *time.Timer
	length int
	modify chan timermodify
}

func (this *timerlist) Len() int {
	return this.length
}

func (this *timerlist) Less(i, j int) bool {
	return this.timers[i].next.Before(this.timers[j].next)
}

func (this *timerlist) Swap(i, j int) {
	this.timers[i], this.timers[j] = this.timers[j], this.timers[i]
	this.timers[i].index = i
	this.timers[j].index = j
}

func (this *timerlist) Push(x any) {
	if this.length == len(this.timers) {
		var items []*timer
		if this.length == 0 {
			items = make([]*timer, 16)
		} else {
			items = make([]*timer, this.length*2)
		}
		copy(items, this.timers)
		this.timers = items
	}
	if item, ok := x.(*timer); ok {
		item.index = this.length
		this.timers[this.length] = item
		this.length++
	}
}

func (this *timerlist) Pop() any {
	if this.length == 0 {
		return nil
	}
	index := this.length - 1
	result := this.timers[index]
	result.index = -1
	this.length--
	return result
}

var timers = timerlist{
	timer:  time.NewTimer(time.Hour * 24),
	modify: make(chan timermodify),
}
var timeronce sync.Once
var cronparser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
var reusablelist = sync.Pool{}

func reusableNew() *reusable {
	result := &reusable{
		timer{
			running: 0,
			index:   -1,
		},
		nil, nil,
	}
	result.executor = ExecFunc(func() error {
		err := result.action()
		reusablelist.Put(result)
		return err
	})
	result.stop = func() {
		result.Stop()
	}
	return result
}
