package utils

import (
	"container/list"
	"time"
)

type Node struct {
	N int64
	T time.Time
}

type Statistics struct {
	l        *list.List
	duration time.Duration
	last     time.Time
	cache    int64
}

func NewStatistics(duration time.Duration) *Statistics {
	if duration < time.Second {
		duration = time.Second * 5
	}
	return &Statistics{
		l:        list.New(),
		duration: duration,
	}
}
func (s *Statistics) Reset() {
	s.cache = 0
	s.last = time.Now()
	s.l.Init()
}
func (s *Statistics) Speed() (speed int64) {
	now := time.Now()
	if !s.last.IsZero() && s.last.Add(time.Millisecond*100).After(now) {
		speed = s.cache
		return
	}

	if s.l.Len() == 0 {
		return
	}

	element := s.l.Front()
	node := element.Value.(Node)
	speed += node.N
	begin := node.T
	for element = element.Next(); element != nil; element = element.Next() {
		node := element.Value.(Node)
		speed += node.N
	}
	div := int64(now.Sub(begin))
	if div < 1 {
		return
	}
	speed *= int64(time.Second)
	speed /= div
	s.cache = speed
	s.last = now
	return
}

func (s *Statistics) Push(n ...int64) {
	s.CheckTimeout()

	var sum int64
	for i := 0; i < len(n); i++ {
		sum += n[i]
	}
	if sum != 0 {
		now := time.Now()
		back := s.l.Back()
		if back != nil {
			node := back.Value.(Node)
			if node.T.Add(time.Second / 10).After(now) {
				s.l.Remove(back)
				node.N += sum
				s.l.PushBack(node)
				return
			}
		}
		s.l.PushBack(Node{
			N: sum,
			T: now,
		})
	}
}

func (s *Statistics) CheckTimeout() {
	l := s.l
	if l.Len() == 0 {
		return
	}

	now := time.Now()
	element := l.Front()
	for element != nil {
		node := element.Value.(Node)
		if node.T.Add(s.duration).After(now) {
			break
		}
		next := element.Next()
		l.Remove(element)
		element = next
	}
}
