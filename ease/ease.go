// Package ease has the timing helpers scenes use to turn "ms since the scene
// started" into animation progress.
package ease

import "math"

// Clamp01 limits x to 0..1.
func Clamp01(x float64) float64 { return math.Max(0, math.Min(1, x)) }

// Progress returns how far t (ms) is through [start, start+dur], as 0..1.
func Progress(t, start, dur int) float64 { return Clamp01(float64(t-start) / float64(dur)) }

// Out is a cubic ease-out: fast start, slow finish.
func Out(x float64) float64 {
	u := 1 - x
	return 1 - u*u*u
}

// InOut is a cubic ease-in-out: slow start and finish.
func InOut(x float64) float64 {
	if x < 0.5 {
		return 4 * x * x * x
	}
	u := -2*x + 2
	return 1 - u*u*u/2
}

// Back overshoots past 1 slightly before settling, like a spring.
func Back(x float64) float64 {
	const c1, c3 = 1.70158, 2.70158
	u := x - 1
	return 1 + c3*u*u*u + c1*u*u
}

// Lerp returns the value x (0..1) of the way from a to b.
func Lerp(a, b, x float64) float64 { return a + (b-a)*x }

// Typed returns the prefix of s typed so far at t ms, one byte every per ms
// after start. s should be ASCII.
func Typed(s string, t, start, per int) string {
	return s[:min(len(s), max(0, (t-start)/per))]
}
