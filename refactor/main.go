package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/goast"
	"github.com/dave/dst/decorator/resolver/guess"
	"golang.org/x/tools/go/packages"
)

const (
	// ifaceNeedle is the name of the interface we are interested in refactoring.
	ifaceNeedle = "Wibbler"
	// pkgNeedle is the name of the package where the interface is defined.
	pkgNeedle = "github.com/charithe/go-ast-refactoring/example/example"
)

func main() {
	// find the absolute path to the example.
	dir, err := filepath.Abs(filepath.Join("..", "example"))
	if err != nil {
		exitOnErr(err)
	}

	// load the packages
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode:  packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedSyntax | packages.NeedName | packages.NeedImports | packages.NeedDeps,
		Dir:   dir,
		Fset:  fset,
		Logf:  log.Printf,
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, "./...")
	exitOnErr(err)

	// find the interface definition
	iface, err := findInterface(pkgs)
	if err != nil {
		exitOnErr(err)
	}

	// create a new decorator with support for resolving imports
	d := decorator.NewDecoratorWithImports(fset, "main", goast.New())

	// rewrite sources
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			rewrite(d, pkg, f, iface)
		}
	}
}

func findInterface(pkgs []*packages.Package) (*types.Interface, error) {
	for _, p := range pkgs {
		if p.PkgPath == pkgNeedle {
			return getInterface(p.Types, ifaceNeedle)
		}

		// look through package imports
		for _, i := range p.Types.Imports() {
			if i.Path() == pkgNeedle {
				return getInterface(i, ifaceNeedle)
			}
		}
	}

	return nil, fmt.Errorf("failed to find %s.%s", pkgNeedle, ifaceNeedle)
}

func getInterface(p *types.Package, name string) (*types.Interface, error) {
	obj := p.Scope().Lookup(name)
	if obj != nil {
		return obj.Type().Underlying().(*types.Interface), nil
	}

	return nil, fmt.Errorf("interface %s does not exist in %s", name, p.Path)
}

func rewrite(d *decorator.Decorator, pkg *packages.Package, f *ast.File, iface *types.Interface) {
	updated := false

	df, err := d.DecorateFile(f)
	exitOnErr(err)

	// traverse the AST.
	dst.Inspect(df, func(node dst.Node) bool {
		if node == nil {
			return false
		}

		// is this a function call with at least a single argument?
		fn, ok := node.(*dst.CallExpr)
		if !ok || len(fn.Args) == 0 {
			return true
		}

		// is it a selctor expression? (e.g. x.method())
		se, ok := fn.Fun.(*dst.SelectorExpr)
		if !ok {
			return true
		}

		// find the type of the selection
		selection, ok := pkg.TypesInfo.Selections[d.Ast.Nodes[se].(*ast.SelectorExpr)]
		if !ok {
			return true
		}

		// is the receiver the correct type?
		if !isInterfaceFunc(iface, selection.Recv(), se.Sel.Name) {
			return true
		}

		// get the first argument to the function
		arg := d.Ast.Nodes[fn.Args[0].(dst.Expr)]
		if arg == nil {
			return true
		}

		// if the first argument is not a context, prepend a context
		argType := pkg.TypesInfo.TypeOf(arg.(ast.Expr))
		if argType != nil && argType.String() != "context.Context" {
			log.Printf("Match found at %s", d.Fset.Position(d.Ast.Nodes[fn].Pos()))
			fn.Args = append([]dst.Expr{&dst.CallExpr{Fun: &dst.Ident{Name: "Background", Path: "context"}}}, fn.Args...)
			updated = true
		}

		return true
	})

	if updated {
		p := d.Fset.Position(f.Pos())
		writeFile(p.Filename, df)
	}
}

func isInterfaceFunc(iface *types.Interface, t types.Type, funcName string) bool {
	if !types.Implements(t, iface) {
		return false
	}

	for i := 0; i < iface.NumMethods(); i++ {
		fn := iface.Method(i)
		if fn.Name() == funcName {
			return true
		}
	}

	return false
}

func writeFile(fileName string, df *dst.File) {
	log.Printf("Writing changes to %s", fileName)

	restorer := decorator.NewRestorerWithImports("main", guess.New())

	out, err := os.Create(fileName)
	if err != nil {
		exitOnErr(fmt.Errorf("failed to open %s for writing", fileName, err))
	}

	defer out.Close()

	exitOnErr(restorer.Fprint(out, df))
}

func exitOnErr(err error) {
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
}
