package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"os"

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

	v := &visitor{
		argcounts: make(map[int]int),
	}

	pkgInfos := program.InitialPackages()
	for _, pkgInfo := range pkgInfos {
		for _, astFile := range pkgInfo.Files {
			file := program.Fset.File(astFile.Pos())
			if file == nil {
				// probably has non parsable code.
				continue
			}
			fullpath := file.Name()
			fmt.Printf("scanning %s...\n", fullpath)
			v.pkgInfo = pkgInfo
			v.fset = program.Fset
			ast.Walk(v, astFile)
		}
	}
	fmt.Printf("DONE\n\n")
	fmt.Printf("functions with N return arguments:\n")
	fmt.Printf("-----------------------------\n")
	fmt.Printf("returns | number of functions\n")
	total := 0
	mult := 0
	for k := range deriveSort(deriveKeys(v.argcounts)) {
		fmt.Printf("%7d | %d\n", k, v.argcounts[k])
		total += v.argcounts[k]
		if k >= 2 {
			mult += v.argcounts[k]
		}
	}
	fmt.Printf("-----------------------------\n")
	fmt.Printf("\n")
	fmt.Printf("total number of functions: %d\n", total)
	fmt.Printf("total number of functions with multiple return parameters: %d\n", mult)
	fmt.Printf("number of functions with 2 return arguments, where the second argument is an error: %d\n", v.errcounts)
	fmt.Printf("\n")
	percent := float64(mult-v.errcounts) / float64(total)
	fmt.Printf("percentage of functions where multiple return parameters are really what we want: %f\n", percent*100)
	fmt.Printf("\n")
}

type visitor struct {
	argcounts map[int]int
	errcounts int

	pkgInfo *loader.PackageInfo
	fset    *token.FileSet
}

func (v *visitor) Visit(node ast.Node) (w ast.Visitor) {
	decl, ok := node.(*ast.FuncDecl)
	if !ok {
		return v
	}
	funcType := decl.Type
	if funcType.Results == nil {
		v.argcounts[0]++
	} else {
		res := funcType.Results.List
		v.argcounts[len(res)]++
		if len(res) == 2 {
			typ := v.pkgInfo.TypeOf(res[1].Type)
			if isError(typ) {
				v.errcounts++
			} else {
				if err := printer.Fprint(os.Stdout, v.fset, node); err != nil {
					panic(err)
				}
				fmt.Printf("\n")
			}
		} else if len(res) > 2 {
			if err := printer.Fprint(os.Stdout, v.fset, node); err != nil {
				panic(err)
			}
			fmt.Printf("\n")
		}
	}
	return v
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

func isError(t types.Type) bool {
	typ, ok := t.(*types.Named)
	if !ok {
		return false
	}
	if typ.Obj().Name() == "error" {
		return true
	}
	for i := 0; i < typ.NumMethods(); i++ {
		meth := typ.Method(i)
		if meth.Name() != "Error" {
			continue
		}
		sig, ok := meth.Type().(*types.Signature)
		if !ok {
			// impossible, but lets check anyway
			continue
		}
		if sig.Params().Len() != 0 {
			continue
		}
		res := sig.Results()
		if res.Len() != 1 {
			continue
		}
		b, ok := res.At(0).Type().(*types.Basic)
		if !ok {
			continue
		}
		if b.Kind() != types.String {
			continue
		}
		return true
	}
	return false
}
