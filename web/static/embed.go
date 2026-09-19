// Package static встраивает CSS в бинарный файл Web Service.
package static

import "embed"

// FS содержит *.css из этого каталога.
//
//go:embed *.css
var FS embed.FS
