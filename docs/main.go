// The fuego-dotclaude documentation site — a Fuego project using the shared
// doc theme. The site keeps only its topbar and sidebar in theme/; everything
// else comes from the doctheme Public pack.
package main

import (
	"fmt"
	"os"

	doctheme "github.com/gofuego/fuego-doctheme"
	"github.com/gofuego/fuego/engine"
	"github.com/gofuego/fuego/parsers/markdown"
)

func main() {
	eng := engine.New()
	eng.Register(markdown.Parser())
	eng.Use(doctheme.Public())

	if err := eng.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "fuego: %v\n", err)
		os.Exit(1)
	}
}
