package templates

import "embed"

//go:embed all:global all:typescript all:swift all:go
var FS embed.FS
