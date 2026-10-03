// Command web writes the promo as an embeddable web page (natively), and is
// the program that page runs (as js/wasm).
package main

import (
	"github.com/lestrrat-go/kavad/examples/promo"
	"github.com/lestrrat-go/kavad/web"
)

func main() { web.Main(promo.Show{}) }
