package embeds

import (
	"embed"
)

//go:embed locales
var I18nFiles embed.FS

//go:embed web
var WebFiles embed.FS
