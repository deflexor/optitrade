package dashboard

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embeddedDist embed.FS

func embeddedAssets() fs.FS {
	sub, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return nil
	}
	return sub
}
