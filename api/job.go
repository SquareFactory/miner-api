package api

import _ "embed"

//go:embed job.tmpl
var JobTemplate string

type Algo struct {
	Algo string
	Pool string
}
