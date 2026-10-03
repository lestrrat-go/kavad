// Package promofsm holds machine "promo": the storyboard of a 30-second
// marketing animation. Each scene is a state; the machine decides the order,
// the captions and the timing (by asking the host to schedule events), and
// the host draws.
//
// This file is the source of truth: edit it directly, or with fsm edit, and
// run fsm check on it after.
package promofsm

import (
	"github.com/lestrrat-go/fsm"
	"github.com/lestrrat-go/fsm/run"
	"github.com/lestrrat-go/fsm/typed"

	hostpkg "github.com/lestrrat-go/kavad/examples/promo/stage"
)

// Host is the type of the value actions call as host.
type Host = hostpkg.Host

// Program compiles the machine, bound to host.
func Program(host Host, opts ...run.Option) (*run.Program, error) {
	def, err := Definition()
	if err != nil {
		return nil, err
	}
	return run.Compile(def, append([]run.Option{run.WithHost(host)}, opts...)...)
}

// Definition returns the machine.
func Definition() (*fsm.Machine, error) {
	return fsm.Define("promo",
		fsm.Doc("A 30-second promo for fsm, drawn with letters and geometric shapes. Scenes are states; timing is scheduled events."),
		fsm.Variables(newGlobals),
		fsm.State("intro",
			typed.Entry(introEntry),
			fsm.On("next", fsm.To("states")),
		),
		fsm.State("states",
			typed.Entry(statesEntry),
			fsm.On("next", fsm.To("events")),
		),
		fsm.State("events",
			typed.Entry(eventsEntry),
			fsm.On("next", fsm.To("snippets")),
			// While the scene runs, a token hops around the graph: a nested
			// machine inside the scene state, driven by its own events.
			fsm.State("idle",
				typed.Entry(idleEntry),
				fsm.On("hop", fsm.To("shop")),
			),
			fsm.State("shop",
				typed.Entry(shopEntry),
				fsm.On("hop", fsm.To("pass")),
			),
			fsm.State("pass",
				typed.Entry(passEntry),
				fsm.On("hop", fsm.To("idle")),
			),
		),
		fsm.State("snippets",
			typed.Entry(snippetsEntry),
			fsm.On("next", fsm.To("verify")),
		),
		fsm.State("verify",
			typed.Entry(verifyEntry),
			fsm.On("next", fsm.To("outro")),
		),
		fsm.State("outro",
			typed.Entry(outroEntry),
			fsm.On("next", fsm.To("end")),
		),
		fsm.Final("end",
			typed.Entry(endEntry),
		),
	)
}

// Globals holds the machine's global variables.
type Globals struct {
	Shown int `json:"Shown"`
}

// Embedding aliases: unexported, so they never collide with variable names.
type globals = Globals

// Scope is what guards and actions see as v in states that see no local
// variables: the globals.
type Scope struct {
	*globals
}

// Load fills the scope from the running machine (see fsm/typed).
func (v *Scope) Load(rt fsm.Runtime) error {
	v.globals, _ = rt.Globals().(*globals)
	return nil
}

// ---- Variables, guards and actions ----

func newGlobals() any {
	return &Globals{}
}

// introEntry is the entry of state "intro".
func introEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Scene("intro")
	v.Shown++
	host.Caption("")
	host.Progress(v.Shown)
	host.Schedule("next", 4000)
	return nil
}

// statesEntry is the entry of state "states".
func statesEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Scene("states")
	v.Shown++
	host.Caption("Describe flows as states.")
	host.Progress(v.Shown)
	host.Schedule("next", 5000)
	return nil
}

// eventsEntry is the entry of state "events".
func eventsEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Scene("events")
	v.Shown++
	host.Caption("Drive them with events.")
	host.Progress(v.Shown)
	host.Schedule("next", 6000)
	return nil
}

// idleEntry is the entry of state "idle".
func idleEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Token("idle")
	host.Schedule("hop", 1400)
	return nil
}

// shopEntry is the entry of state "shop".
func shopEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Token("shop")
	host.Schedule("hop", 1400)
	return nil
}

// passEntry is the entry of state "pass".
func passEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Token("pass")
	host.Schedule("hop", 1400)
	return nil
}

// snippetsEntry is the entry of state "snippets".
func snippetsEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Scene("snippets")
	v.Shown++
	host.Caption("Logic is plain Go.")
	host.Progress(v.Shown)
	host.Schedule("next", 5000)
	return nil
}

// verifyEntry is the entry of state "verify".
func verifyEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Scene("verify")
	v.Shown++
	host.Caption("Checked before it ships.")
	host.Progress(v.Shown)
	host.Schedule("next", 5000)
	return nil
}

// outroEntry is the entry of state "outro".
func outroEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Scene("outro")
	v.Shown++
	host.Caption("State machines — written by agents, read by humans.")
	host.Progress(v.Shown)
	host.Schedule("next", 5000)
	return nil
}

// endEntry is the entry of state "end".
func endEntry(v Scope, host Host, rt fsm.Runtime) error {
	host.Scene("end")
	return nil
}
