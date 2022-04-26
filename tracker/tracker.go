package tracker

import (
	"time"
)

type (
	Tracker struct {
		base time.Time
		skip int64
	}
)

//sec:60 -> 60secs (not 60 nanosec!)
func (t *Tracker) Init(sec int64) *Tracker {
	uTime := time.Now().Unix() //time.Time -> Unix time
	mod := uTime % sec
	bUnix := uTime - mod         //starting point
	bTime := time.Unix(bUnix, 0) // Unix time -> time.Time
	t.base = bTime
	t.skip = sec
	return t
}

/*数ミリ秒ずつ増えていってしまう。Unix時間に変換するとミリ秒以下は落ちてしまうため
微妙に誤差が積みあがってしまうのかも。
func (t *Tracker) Wait() *Tracker {
	next := t.base.Unix() + t.skip // base time + skip in Unix time
	now := time.Now().Unix()
	diff := next - now
	time.Sleep(time.Duration(diff) * time.Second)
	t.base = time.Unix(next, 0)
	fmt.Printf("base:%v\n", t.base)
	return t
}
*/

//こっちはずれない
func (t *Tracker) Wait() *Tracker {
	next := t.base.Add(time.Duration(t.skip) * time.Second) //next = current + skip
	diff := next.Sub(time.Now())                            //next - current time
	time.Sleep(time.Duration(diff))
	t.base = next
	return t
}

//constructor
func NewTracker(sec int64) *Tracker {
	return (&Tracker{}).Init(sec)
}
