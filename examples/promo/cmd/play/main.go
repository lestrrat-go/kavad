// Command play shows the promo in a native window.
package main

import (
	"github.com/lestrrat-go/kavad/desktop"
	"github.com/lestrrat-go/kavad/examples/promo"
)

func main() { desktop.Main(promo.Show{}) }
