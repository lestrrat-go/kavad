package main

import (
	"github.com/lestrrat-go/kavad/examples/imageassets"
	"github.com/lestrrat-go/kavad/web"
)

func main() { web.Main(imageassets.Show{}) }
