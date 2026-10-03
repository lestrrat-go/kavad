package kavad_test

import (
	"testing"

	"github.com/lestrrat-go/kavad"
	"github.com/stretchr/testify/require"
)

type delivery struct {
	at    int
	event string
}

func TestStageRunUntil(t *testing.T) {
	st := kavad.NewStage()
	st.Schedule("b", 200)
	st.Schedule("a", 100)
	st.Schedule("c", 200) // same time as "b": delivered after it

	var got []delivery
	deliver := func(event string) error {
		got = append(got, delivery{st.Now(), event})
		return nil
	}
	require.NoError(t, st.RunUntil(150, deliver))
	require.Equal(t, []delivery{{100, "a"}}, got)
	require.Equal(t, 150, st.Now())

	require.NoError(t, st.RunUntil(1000, deliver))
	require.Equal(t, []delivery{{100, "a"}, {200, "b"}, {200, "c"}}, got)
	require.Equal(t, 1000, st.Now())
}

// Calls made while an event is delivered are stamped with the event's due
// time, not with the time RunUntil was asked to reach.
func TestStageStampsEventTime(t *testing.T) {
	st := kavad.NewStage()
	st.Scene("one")
	st.Schedule("next", 300)
	require.NoError(t, st.RunUntil(1000, func(event string) error {
		switch event {
		case "next":
			st.Scene("two")
			st.Schedule("last", 50)
		case "last":
			st.Scene("three")
		}
		return nil
	}))
	require.Equal(t, []kavad.Change{{At: 0, Scene: "one"}, {At: 300, Scene: "two"}, {At: 350, Scene: "three"}}, st.History())
	scene, at := st.Current()
	require.Equal(t, "three", scene)
	require.Equal(t, 350, at)
}

func TestParseHex(t *testing.T) {
	c, err := kavad.ParseHex("#ff6b35")
	require.NoError(t, err)
	require.Equal(t, kavad.Color{R: 0xff, G: 0x6b, B: 0x35, A: 1}, c)
	require.Equal(t, "#ff6b35", c.Hex())

	for _, bad := range []string{"", "ff6b35", "#ff6b3", "#gg0000"} {
		_, err := kavad.ParseHex(bad)
		require.Error(t, err, bad)
	}
}
