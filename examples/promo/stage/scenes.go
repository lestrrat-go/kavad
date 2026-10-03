package stage

import (
	"math"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/draw"
	"github.com/lestrrat-go/kavad/ease"
)

// monoAdvance is the width of one monospace character at the given size.
func monoAdvance(size float64) float64 { return size * 0.6 }

// Paint implements kavad.Painter: it draws the picture at time now (ms).
func (s *Screen) Paint(c kavad.Canvas, now int) {
	p := draw.New(c)
	p.Rect(0, 0, W, H, colBG)
	scene, sceneAt := s.Current()
	t := now - sceneAt
	grid(p, now)
	switch scene {
	case "intro":
		intro(p, t)
	case "states":
		states(p, t)
	case "events":
		s.events(p, now)
	case "snippets":
		snippets(p, now, t)
	case "verify":
		verify(p, t)
	case "outro":
		outro(p, now, t)
	}
	s.captionLayer(p, now)
	s.progressLayer(p)
}

// grid draws the faint dot grid behind every scene; it fades in during the
// first second of the video.
func grid(p draw.Pen, now int) {
	p = p.Fade(0.18 * ease.Progress(now, 0, 900))
	for y := 40; y < H; y += 40 {
		for x := 40; x < W; x += 40 {
			p.Circle(float64(x), float64(y), 1.6, colDim)
		}
	}
}

// shape draws a geometric shape centred on (x, y) with size r, rotated by
// rot degrees.
func shape(p draw.Pen, kind string, x, y, r, rot float64, c kavad.Color) {
	switch kind {
	case "circle":
		p.Circle(x, y, r, c)
	case "square":
		p.Fill(draw.Rotated(x, y, rot, kavad.Point{X: -r, Y: -r}, kavad.Point{X: r, Y: -r}, kavad.Point{X: r, Y: r}, kavad.Point{X: -r, Y: r}), c)
	case "triangle":
		p.Fill(draw.Rotated(x, y, rot, kavad.Point{X: 0, Y: -r * 1.15}, kavad.Point{X: r, Y: r * 0.75}, kavad.Point{X: -r, Y: r * 0.75}), c)
	}
}

// ---- intro: shapes become the logo ----

// letterX centres each logo letter; the glyphs differ in width, so even
// spacing would leave a gap after the narrow "f".
var letterX = [3]float64{496, 604, 746}

var logo = []struct {
	letter, shape string
	color         kavad.Color
	fromX, fromY  float64
}{
	{"f", "circle", colAccent, 170, 140},
	{"s", "square", colTeal, 640, 640},
	{"m", "triangle", colYellow, 1110, 140},
}

// letters draws the "fsm" logo at baseline y, each letter at the given
// progress (0 = absent, 1 = fully there).
func letters(p draw.Pen, y, size float64, lp [3]float64) {
	for i, l := range logo {
		sc := ease.Lerp(0.6, 1, ease.Back(lp[i]))
		p.Fade(lp[i]).Text(kavad.Text{X: letterX[i], Y: y, S: l.letter, Font: kavad.SansBold, Size: size * sc, Color: colFG, Align: kavad.AlignCenter})
	}
}

func underline(p draw.Pen, y, u float64) {
	if u > 0 {
		p.Fill(draw.RoundRect(640-180*u, y, 360*u, 8, math.Min(4, 180*u)), colAccent)
	}
}

func intro(p draw.Pen, t int) {
	var lp [3]float64
	for i, l := range logo {
		start := 200 + i*180
		move := ease.Out(ease.Progress(t, start, 1000))
		x, y := ease.Lerp(l.fromX, letterX[i], move), ease.Lerp(l.fromY, 330, move)
		morph := ease.Progress(t, start+900, 500)
		shape(p.Fade(1-morph), l.shape, x, y, ease.Lerp(46, 20, morph), ease.Lerp(-90, 0, move), l.color)
		lp[i] = morph
	}
	letters(p, 385, 170, lp)
	underline(p, 420, ease.Out(ease.Progress(t, 2200, 600)))
	p.Fade(ease.Progress(t, 2700, 600)).Text(kavad.Text{X: 640, Y: 490, S: "STATE MACHINES FOR GO", Font: kavad.Sans, Size: 30, Color: colFG, Align: kavad.AlignCenter, Spacing: 6})
}

// ---- states and events: the graph ----

// The promo machine's state names, as the graph scenes draw them.
const (
	stIdle = "idle"
	stShop = "shop"
	stPass = "pass"
)

var nodes = map[string]kavad.Point{
	stIdle: {X: 330, Y: 330},
	stShop: {X: 640, Y: 230},
	stPass: {X: 950, Y: 330},
}

var nodeOrder = []string{stIdle, stShop, stPass}

type edge struct {
	from, to, label string
	bend            float64 // control point offset for curves
}

var edges = []edge{
	{stIdle, stShop, "talk", 0},
	{stShop, stPass, "buy", 0},
	{stPass, stIdle, "loop", 170},
}

const nodeR = 62.0

// edgePath returns the polyline for e between the circles' edges, the arrow
// tip and its direction, and the control point.
func edgePath(e edge) ([]kavad.Point, kavad.Point, kavad.Point, kavad.Point) {
	a, b := nodes[e.from], nodes[e.to]
	m := kavad.Point{X: (a.X + b.X) / 2, Y: (a.Y+b.Y)/2 + e.bend}
	// Start/end directions: towards the control point for curves.
	dir := func(f, t kavad.Point) kavad.Point {
		l := math.Hypot(t.X-f.X, t.Y-f.Y)
		return kavad.Point{X: (t.X - f.X) / l, Y: (t.Y - f.Y) / l}
	}
	sd, ed := dir(a, m), dir(b, m)
	start := kavad.Point{X: a.X + sd.X*nodeR, Y: a.Y + sd.Y*nodeR}
	tip := kavad.Point{X: b.X + ed.X*(nodeR+10), Y: b.Y + ed.Y*(nodeR+10)}
	pts := []kavad.Point{start, tip}
	if e.bend != 0 {
		pts = draw.Quad(start, m, tip)
	}
	return pts, tip, kavad.Point{X: -ed.X, Y: -ed.Y}, m
}

func drawEdge(p draw.Pen, e edge, drawn, labelOpacity float64, c kavad.Color) {
	pts, tip, d, m := edgePath(e)
	p.Stroke(draw.Trim(pts, drawn), 4, c)
	if drawn > 0.97 {
		p.Fill(draw.Arrowhead(tip.X, tip.Y, d.X, d.Y), c)
	}
	lx, ly := m.X, m.Y-18
	if e.bend != 0 {
		ly = m.Y - e.bend/2 + 40
	}
	p.Fade(labelOpacity).Text(kavad.Text{X: lx, Y: ly, S: e.label, Font: kavad.Mono, Size: 22, Color: colLabel, Align: kavad.AlignCenter})
}

func drawNode(p draw.Pen, id string, scale float64, stroke kavad.Color) {
	if scale <= 0 {
		return
	}
	c := nodes[id]
	p.Circle(c.X, c.Y, nodeR*scale, colBG)
	p.Ring(c.X, c.Y, nodeR*scale, 4*scale, stroke)
	p.Text(kavad.Text{X: c.X, Y: c.Y + 9*scale, S: id, Font: kavad.Sans, Size: 26 * scale, Color: colFG, Align: kavad.AlignCenter})
}

func states(p draw.Pen, t int) {
	for k, e := range edges {
		drawEdge(p, e, ease.InOut(ease.Progress(t, 1300+k*450, 600)), ease.Progress(t, 1800+k*450, 400), colDim)
	}
	for i, id := range nodeOrder {
		drawNode(p, id, ease.Back(ease.Progress(t, 200+i*250, 500)), colFG)
	}
}

func (s *Screen) events(p draw.Pen, now int) {
	moving := ""
	for _, e := range edges {
		col := colDim
		if e.from == s.prev && e.to == s.token && now-s.tokenAt < 900 {
			col, moving = colAccent, e.label
		}
		drawEdge(p, e, 1, 1, col)
	}
	for _, id := range nodeOrder {
		stroke := colFG
		if id == s.token {
			stroke = colAccent
		}
		drawNode(p, id, 1, stroke)
	}
	// The token travels from the previous state to the current one.
	if s.token != "" {
		at := nodes[s.token]
		if s.prev != "" {
			from := nodes[s.prev]
			k := ease.InOut(ease.Progress(now, s.tokenAt, 600))
			at = kavad.Point{X: ease.Lerp(from.X, at.X, k), Y: ease.Lerp(from.Y, at.Y, k)}
			if s.prev == stPass && s.token == stIdle { // follow the curve below
				at.Y += 4 * k * (1 - k) * 170
			}
		}
		pulse := 1 + 0.12*math.Sin(float64(now)/120)
		p.Circle(at.X, at.Y-nodeR-26, 13*pulse, colAccent)
		p.Stroke([]kavad.Point{{X: at.X, Y: at.Y - nodeR - 13}, {X: at.X, Y: at.Y - nodeR}}, 3, colAccent)
	}
	// The event chip at the top.
	if moving != "" {
		chip := p.Fade(1 - ease.Progress(now, s.tokenAt+500, 400))
		chip.Fill(draw.RoundRect(540, 70, 200, 48, 24), colAccent)
		chip.Text(kavad.Text{X: 640, Y: 102, S: "event: " + moving, Font: kavad.Mono, Size: 24, Color: colBG, Align: kavad.AlignCenter})
	}
}

// ---- snippets: the code card ----

type seg struct {
	text  string
	color kavad.Color
}

var code = [][]seg{
	{{"on ", colAccent}, {`"buy"`, colYellow}},
	{{"if ", colAccent}, {"v.Gold >= v.Price", colFG}},
	{{"do ", colAccent}, {"v.Gold -= v.Price", colFG}},
	{{"   host.Say(", colFG}, {`"A fine blade."`, colYellow}, {")", colFG}},
	{{"to ", colAccent}, {`"pass"`, colYellow}},
}

func floaters(p draw.Pen, now int) {
	kinds := []string{"square", "triangle", "circle", "square", "triangle"}
	for i, k := range kinds {
		x := 120 + float64(i)*260 + 30*math.Sin(float64(now)/1500+float64(i))
		y := 110 + float64(i%2)*470 + 20*math.Cos(float64(now)/1300+float64(i)*2)
		shape(p.Fade(0.35), k, x, y, 22, float64(now)/40+float64(i)*30, colDim)
	}
}

func snippets(p draw.Pen, now, t int) {
	floaters(p.Fade(ease.Progress(t, 0, 600)), now)
	in := ease.Out(ease.Progress(t, 0, 500))
	card := p.Fade(in).Shift(0, 20*(1-in))
	card.Fill(draw.RoundRect(300, 150, 680, 360, 18), colPanel)
	card.StrokeClosed(draw.RoundRect(300, 150, 680, 360, 18), 2, colDim)
	for i, c := range []kavad.Color{colRed, colYellow, colGreen} {
		card.Circle(330+float64(i)*24, 180, 7, c)
	}
	card.Text(kavad.Text{X: 960, Y: 187, S: "shop.on[2]", Font: kavad.Mono, Size: 18, Color: colDim, Align: kavad.AlignEnd})

	// Type the snippet out, 32 ms per character.
	const size = 28.0
	left := max(0, (t-600)/32)
	cx, cy := 340.0, 250.0
	for li, line := range code {
		if left <= 0 {
			break
		}
		y := 250 + float64(li)*48
		col := 0
		for _, sg := range line {
			n := min(len(sg.text), left)
			p.Text(kavad.Text{X: 340 + float64(col)*monoAdvance(size), Y: y, S: sg.text[:n], Font: kavad.Mono, Size: size, Color: sg.color})
			left -= n
			col += n
			if left <= 0 {
				break
			}
		}
		cx, cy = 340+float64(col)*monoAdvance(size), y
	}
	if (now/450)%2 == 0 {
		p.Fade(0.85).Rect(cx+2, cy-24, 15, 30, colAccent)
	}
}

// ---- verify: fsm check finds and clears a dead end ----

var vnodes = map[string]kavad.Point{stIdle: {X: 380, Y: 290}, stShop: {X: 640, Y: 290}, "lost": {X: 900, Y: 290}}

func vnode(p draw.Pen, id string, stroke kavad.Color) {
	c := vnodes[id]
	p.Circle(c.X, c.Y, 52, colBG)
	p.Ring(c.X, c.Y, 52, 4, stroke)
	p.Text(kavad.Text{X: c.X, Y: c.Y + 8, S: id, Font: kavad.Sans, Size: 24, Color: colFG, Align: kavad.AlignCenter})
}

func verify(p draw.Pen, t int) {
	p = p.Fade(ease.Progress(t, 0, 400))
	p.Stroke([]kavad.Point{{X: 434, Y: 290}, {X: 576, Y: 290}}, 4, colDim)
	p.Stroke([]kavad.Point{{X: 694, Y: 290}, {X: 836, Y: 290}}, 4, colDim)
	p.Fill([]kavad.Point{{X: 586, Y: 290}, {X: 572, Y: 281}, {X: 572, Y: 299}}, colDim)
	p.Fill([]kavad.Point{{X: 846, Y: 290}, {X: 832, Y: 281}, {X: 832, Y: 299}}, colDim)
	// The fix: an arrow from lost back to idle, drawn over the top.
	if fixed := ease.Progress(t, 2800, 600); fixed > 0 {
		p.Stroke(draw.Trim(draw.Quad(kavad.Point{X: 900, Y: 238}, kavad.Point{X: 640, Y: 90}, kavad.Point{X: 392, Y: 236}), ease.InOut(fixed)), 4, colGreen)
		if fixed > 0.97 {
			p.Fill([]kavad.Point{{X: 388, Y: 240}, {X: 384, Y: 224}, {X: 398, Y: 232}}, colGreen)
		}
	}
	broken := ease.Progress(t, 1400, 300)
	healed := ease.Progress(t, 3500, 400)
	lostStroke := colFG
	if broken > 0 {
		lostStroke = colRed
	}
	if healed > 0 {
		lostStroke = colGreen
	}
	vnode(p, stIdle, colFG)
	vnode(p, stShop, colFG)
	vnode(p, "lost", lostStroke)
	// Badge: an X while broken, a check once healed.
	bx, by := 945.0, 245.0
	if broken > 0 && healed < 1 {
		r := 18 * ease.Back(broken) * (1 - healed)
		p.Circle(bx, by, r, colRed)
		p.Stroke([]kavad.Point{{X: bx - r*0.4, Y: by - r*0.4}, {X: bx + r*0.4, Y: by + r*0.4}}, 4, colBG)
		p.Stroke([]kavad.Point{{X: bx - r*0.4, Y: by + r*0.4}, {X: bx + r*0.4, Y: by - r*0.4}}, 4, colBG)
	}
	if healed > 0 {
		r := 18 * ease.Back(healed)
		p.Circle(bx, by, r, colGreen)
		p.Stroke([]kavad.Point{{X: bx - r*0.45, Y: by}, {X: bx - r*0.15, Y: by + r*0.35}, {X: bx + r*0.4, Y: by - r*0.35}}, 4, colBG)
	}
	// The terminal.
	p.Fill(draw.RoundRect(290, 420, 700, 130, 14), colPanel)
	p.StrokeClosed(draw.RoundRect(290, 420, 700, 130, 14), 2, colDim)
	p.Text(kavad.Text{X: 320, Y: 465, S: ease.Typed("$ fsm check quest.fsm.json", t, 200, 35), Font: kavad.Mono, Size: 24, Color: colFG})
	p.Fade(ease.Progress(t, 1400, 250) * (1 - ease.Progress(t, 3500, 250))).Text(kavad.Text{X: 320, Y: 515, S: `error  dead-end  state "lost" has no way out`, Font: kavad.Mono, Size: 22, Color: colRed})
	p.Fade(ease.Progress(t, 3700, 300)).Text(kavad.Text{X: 320, Y: 515, S: "quest.fsm.json: ok", Font: kavad.Mono, Size: 22, Color: colGreen})
}

// ---- outro ----

func outro(p draw.Pen, now, t int) {
	// Shapes orbiting the logo.
	for i, l := range logo {
		a := float64(now)/1400 + float64(i)*2*math.Pi/3
		x, y := 640+420*math.Cos(a), 310+210*math.Sin(a)
		shape(p.Fade(ease.Progress(t, 300+i*150, 500)), l.shape, x, y, 20, float64(now)/25, l.color)
	}
	var lp [3]float64
	for i := range lp {
		lp[i] = ease.Progress(t, i*130, 450)
	}
	letters(p, 340, 170, lp)
	underline(p, 375, ease.Out(ease.Progress(t, 600, 500)))
	// Left-anchored where the full URL would start centred, so the text does
	// not slide while it is typed.
	const url, size = "github.com/lestrrat-go/fsm", 30.0
	p.Text(kavad.Text{X: 640 - float64(len(url))*monoAdvance(size)/2, Y: 450, S: ease.Typed(url, t, 1200, 40), Font: kavad.Mono, Size: size, Color: colTeal})
	// Fade to black at the very end.
	p.Fade(ease.Progress(t, 4400, 600)).Rect(0, 0, W, H, colBG)
}

// ---- overlays ----

func (s *Screen) captionLayer(p draw.Pen, now int) {
	if s.caption == "" {
		return
	}
	k := ease.Out(ease.Progress(now, s.capAt+250, 500))
	if scene, sceneAt := s.Current(); scene == "outro" {
		p = p.Fade(1 - ease.Progress(now-sceneAt, 4400, 600))
	}
	p.Fade(k).Text(kavad.Text{X: 640, Y: 630 + 12*(1-k), S: s.caption, Font: kavad.Sans, Size: 34, Color: colFG, Align: kavad.AlignCenter})
}

func (s *Screen) progressLayer(p draw.Pen) {
	if scene, _ := s.Current(); scene == "end" {
		return
	}
	for i := range 6 {
		x := 1150 + float64(i)*18
		if i < s.progress {
			p.Circle(x, 686, 5, colFG)
		} else {
			p.Ring(x, 686, 4.5, 1.5, colDim)
		}
	}
}
