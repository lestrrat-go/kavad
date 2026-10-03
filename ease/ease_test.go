package ease_test

import (
	"testing"

	"github.com/lestrrat-go/kavad/ease"
	"github.com/stretchr/testify/require"
)

func TestProgress(t *testing.T) {
	require.Equal(t, 0.0, ease.Progress(50, 100, 200))
	require.Equal(t, 0.5, ease.Progress(200, 100, 200))
	require.Equal(t, 1.0, ease.Progress(900, 100, 200))
}

func TestCurvesHitEnds(t *testing.T) {
	for name, f := range map[string]func(float64) float64{"Out": ease.Out, "InOut": ease.InOut, "Back": ease.Back} {
		require.InDelta(t, 0, f(0), 1e-9, name)
		require.InDelta(t, 1, f(1), 1e-9, name)
	}
	require.Greater(t, ease.Back(0.8), 1.0) // overshoots before settling
}

func TestTyped(t *testing.T) {
	require.Equal(t, "", ease.Typed("hello", 100, 200, 10))
	require.Equal(t, "hel", ease.Typed("hello", 230, 200, 10))
	require.Equal(t, "hello", ease.Typed("hello", 9999, 200, 10))
}
