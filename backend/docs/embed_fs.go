package docs

import "embed"

//go:embed swagger.json
var Docs embed.FS
