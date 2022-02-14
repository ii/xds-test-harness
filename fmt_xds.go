package main

import (
	// "fmt"
	"io"
	// "math"

	"github.com/cucumber/godog"
	// "github.com/rs/zerolog/log"
)

func init() {
	godog.Format("xds", "Formatter for the xDS Test Suite", xdsFormatterFunc)
}

func xdsFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newXdsFmt(suite, out)
}

func newXdsFmt(suite string, out io.Writer) *xdsFmt {
	return &xdsFmt{
		CukeFmt: godog.NewCukeFmt(suite, out),
		out:      out,
	}
}

type xdsFmt struct {
	*godog.CukeFmt
	out io.Writer
}
