package kavad

import "sort"

// Host is the part of a show's host interface that Stage implements. Embed
// it in the show's own host interface (the Host type of its machine file) so
// actions can call host.Scene and host.Schedule.
type Host interface {
	// Scene switches the picture to the named scene and restarts its
	// animation.
	Scene(name string)
	// Schedule asks for event to be sent to the machine after ms
	// milliseconds.
	Schedule(event string, ms int)
}

// Change records a Scene call.
type Change struct {
	At    int // ms
	Scene string
}

type scheduled struct {
	due, seq int
	event    string
}

// Stage is the clock-free core of a host: the current scene and the queue of
// scheduled events. Embed *Stage in a show's host type.
//
// Stage does not own a clock. Its runner calls RunUntil with the current
// time, which delivers due events to the machine; the machine's actions
// then call Scene and Schedule, which Stage stamps with that time.
type Stage struct {
	now     int
	scene   string
	sceneAt int
	queue   []scheduled
	seq     int
	history []Change
}

var _ Host = (*Stage)(nil)

// NewStage returns a stage at time 0 with no scene.
func NewStage() *Stage { return &Stage{} }

// Scene implements Host.
func (s *Stage) Scene(name string) {
	s.scene, s.sceneAt = name, s.now
	s.history = append(s.history, Change{At: s.now, Scene: name})
}

// Schedule implements Host. Events due at the same time are delivered in the
// order they were scheduled.
func (s *Stage) Schedule(event string, ms int) {
	s.seq++
	s.queue = append(s.queue, scheduled{due: s.now + ms, seq: s.seq, event: event})
	sort.SliceStable(s.queue, func(i, j int) bool {
		if s.queue[i].due != s.queue[j].due {
			return s.queue[i].due < s.queue[j].due
		}
		return s.queue[i].seq < s.queue[j].seq
	})
}

// Now returns the stage's current time in ms.
func (s *Stage) Now() int { return s.now }

// Current returns the current scene and the time it started.
func (s *Stage) Current() (string, int) { return s.scene, s.sceneAt }

// History returns every Scene call so far.
func (s *Stage) History() []Change { return append([]Change(nil), s.history...) }

// RunUntil advances the clock to now, delivering every event scheduled at or
// before it, in order. The clock is set to each event's due time before it is
// delivered, so whatever the event starts is stamped with the scheduled time,
// not the frame boundary.
func (s *Stage) RunUntil(now int, deliver func(event string) error) error {
	for len(s.queue) > 0 && s.queue[0].due <= now {
		next := s.queue[0]
		s.queue = s.queue[1:]
		s.now = next.due
		if err := deliver(next.event); err != nil {
			return err
		}
	}
	s.now = now
	return nil
}
