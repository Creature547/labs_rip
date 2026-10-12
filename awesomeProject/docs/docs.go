// Package docs registers the bundled Swagger specification.
// It can be regenerated using swag init -g cmd/awesomeProject/main.go.
package docs

import (
	_ "embed"
	"github.com/swaggo/swag"
)

//go:embed swagger.json
var specification string

type document struct{}

func (d *document) ReadDoc() string { return specification }
func init()                         { swag.Register(swag.Name, &document{}) }
