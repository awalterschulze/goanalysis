package main

import (
	"flag"
	"fmt"
	"go/parser"

	"github.com/kisielk/gotool"
	"golang.org/x/tools/go/loader"
)

func main() {
	flag.Parse()
	paths := gotool.ImportPaths(flag.Args())
	program, err := load(paths...)
	if err != nil {
		panic(err)
	}
	_ = program
}

func load(paths ...string) (*loader.Program, error) {
	conf := loader.Config{
		ParserMode:  parser.ParseComments,
		AllowErrors: true,
	}
	conf.TypeChecker.Error = func(err error) {}
	rest, err := conf.FromArgs(paths, true)
	if err != nil {
		return nil, fmt.Errorf("could not parse arguments: %s", err)
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("unhandled extra arguments: %v", rest)
	}
	p, err := conf.Load()
	if err != nil {
		return nil, err
	}
	if p.Fset == nil {
		return nil, fmt.Errorf("program == nil")
	}
	return p, nil
}
