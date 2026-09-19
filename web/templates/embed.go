// Package templates встраивает HTML-шаблоны в бинарный файл Web Service.
package templates

import "embed"

// FS содержит *.html из этого каталога.
//
//go:embed *.html
var FS embed.FS
