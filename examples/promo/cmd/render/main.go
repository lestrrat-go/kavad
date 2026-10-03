// Command render writes the promo frame by frame into out/frames.html.
package main

import (
	"github.com/lestrrat-go/kavad/examples/promo"
	"github.com/lestrrat-go/kavad/svg"
)

func main() { svg.Main(promo.Show{}) }
