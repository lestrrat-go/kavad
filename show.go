package kavad

import "context"

// Show is a complete show, ready to play any number of times.
type Show interface {
	// Size returns the show's canvas size. Renderers scale it to their
	// output and keep its aspect ratio.
	Size() (w, h int)
	// Start begins a fresh run at time 0: a new host, a new machine
	// instance in its initial state.
	Start(ctx context.Context) (*Run, error)
}

// Machine is the part of an fsm instance a Run drives. *run.Instance
// implements it.
type Machine interface {
	Send(ctx context.Context, event string) error
	Done() bool
}

// Painter draws a show at a moment. A show's host implements it.
type Painter interface {
	Paint(c Canvas, now int)
}

// Run is one playthrough of a show: a machine instance, the stage its host
// embeds, and the painter that draws it.
type Run struct {
	m       Machine
	stage   *Stage
	painter Painter
	images  []*Image
}

// NewRun ties a machine instance to its host's stage and painter, and lists
// every image the painter might draw. None of the arguments may be nil.
func NewRun(m Machine, stage *Stage, painter Painter, images ...*Image) *Run {
	for _, img := range images {
		if img == nil {
			panic("kavad: nil run image")
		}
	}
	return &Run{m: m, stage: stage, painter: painter, images: append([]*Image(nil), images...)}
}

// Images returns the images declared for this run. Runners prepare them before
// drawing, so the show can choose among them without loading on each frame.
func (r *Run) Images() []*Image { return append([]*Image(nil), r.images...) }

// Advance moves the run to now (ms since it started), sending the machine
// every event scheduled up to then. Events due after the machine finished
// are dropped.
func (r *Run) Advance(ctx context.Context, now int) error {
	return r.stage.RunUntil(now, func(event string) error {
		if r.m.Done() {
			return nil
		}
		return r.m.Send(ctx, event)
	})
}

// Draw paints the moment now on c.
func (r *Run) Draw(c Canvas, now int) { r.painter.Paint(c, now) }

// Done reports whether the machine reached a final state.
func (r *Run) Done() bool { return r.m.Done() }

// Stage returns the run's stage.
func (r *Run) Stage() *Stage { return r.stage }

// Loop plays a show over and over, signage style: when a run finishes, it
// waits, then starts a fresh one. Runners feed it their clock with Update and
// draw with Draw.
type Loop struct {
	show   Show
	hold   int
	run    *Run
	start  int // clock value at which the current run started
	now    int // ms into the current run
	doneAt int // ms into the run at which it finished, or -1
}

// NewLoop starts show. After a run finishes, the loop keeps drawing its last
// moment for hold ms, then restarts; a negative hold never restarts.
func NewLoop(ctx context.Context, show Show, hold int) (*Loop, error) {
	l := &Loop{show: show, hold: hold}
	if err := l.restart(ctx, 0); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Loop) restart(ctx context.Context, clock int) error {
	r, err := l.show.Start(ctx)
	if err != nil {
		return err
	}
	l.run, l.start, l.now, l.doneAt = r, clock, 0, -1
	return nil
}

// Update moves the loop to clock, the runner's time in ms. clock must not go
// backwards.
func (l *Loop) Update(ctx context.Context, clock int) error {
	now := clock - l.start
	if err := l.run.Advance(ctx, now); err != nil {
		return err
	}
	if l.run.Done() && l.doneAt < 0 {
		l.doneAt = now
	}
	if l.hold >= 0 && l.doneAt >= 0 && now-l.doneAt >= l.hold {
		return l.restart(ctx, clock)
	}
	l.now = now
	return nil
}

// Draw paints the current moment on c.
func (l *Loop) Draw(c Canvas) { l.run.Draw(c, l.now) }

// Now returns how far the current run is, in ms.
func (l *Loop) Now() int { return l.now }

// Run returns the current run.
func (l *Loop) Run() *Run { return l.run }
