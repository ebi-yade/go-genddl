package genddl

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Field struct {
	ColumnDef string
}

type Table struct {
	Name       string
	Fields     []Field
	PrimaryKey string
}

func Run(from string) {
	fromdir := filepath.Dir(from)

	var schemadir, outpath, driverName string
	flag.StringVar(&schemadir, "schemadir", fromdir, "schema declaretion directory")
	flag.StringVar(&outpath, "outpath", "", "schema target path")
	flag.StringVar(&driverName, "driver", "mysql", "target driver")

	flag.Parse()

	var dialect Dialect
	switch driverName {
	case "mysql":
		dialect = MysqlDialect{}
	case "sqlite3":
		dialect = Sqlite3Dialect{}
	default:
		log.Fatalf("undefined driver name: %s", driverName)
	}

	tables, funcMap, err := retrieveTables(schemadir)
	if err != nil {
		log.Fatal("parse and retrieve table error: %s", err)
	}

	file, err := os.Create(outpath)
	if err != nil {
		log.Fatal("invalid outpath error:", err)
	}
	tablesMap := map[*ast.StructType]string{}
	var tableNames []string
	for tableName, st := range tables {
		tablesMap[st] = tableName
		tableNames = append(tableNames, tableName)
	}
	sort.Strings(tableNames)
	file.WriteString("-- generated by github.com/mackee/go-genddl. DO NOT EDIT!!!\n")
	for _, tableName := range tableNames {
		st := tables[tableName]
		funcs := funcMap[st]
		tableMap := NewTableMap(tableName, st, funcs, tablesMap)
		if tableMap != nil {
			file.WriteString("\n")
			tableMap.WriteDDL(file, dialect)
		}
	}

}

var typeNameStructMap = map[string]*ast.StructType{}

func retrieveTables(schemadir string) (map[string]*ast.StructType, map[*ast.StructType][]*ast.FuncDecl, error) {
	path, err := filepath.Abs(schemadir)
	if err != nil {
		return nil, nil, err
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(
		fset,
		path,
		func(finfo os.FileInfo) bool { return true },
		parser.ParseComments,
	)

	if err != nil {
		return nil, nil, err
	}

	var decls []ast.Decl
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			decls = append(decls, file.Decls...)
		}
	}

	tables := map[string]*ast.StructType{}
	funcs := []*ast.FuncDecl{}
	funcMap := map[*ast.StructType][]*ast.FuncDecl{}
	for _, decl := range decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			funcs = append(funcs, funcDecl)
		}
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Doc == nil {
				continue
			}

			comment := genDecl.Doc.List[0]
			if strings.HasPrefix(comment.Text, "//+table:") {
				tableName := strings.TrimPrefix(comment.Text, "//+table:")
				tableName = strings.TrimSpace(tableName)
				spec := genDecl.Specs[0]
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}

				tables[tableName] = st
				funcMap[st] = make([]*ast.FuncDecl, 0)
				typeNameStructMap[ts.Name.Name] = st
			}
		}
	}

	for _, funcDecl := range funcs {
		if funcDecl.Recv.NumFields() < 0 {
			continue
		}
		recv := funcDecl.Recv.List[0]
		ident, ok := recv.Type.(*ast.Ident)
		if !ok {
			continue
		}
		ts, ok := ident.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			continue
		}
		if funcs, ok := funcMap[st]; ok {
			funcMap[st] = append(funcs, funcDecl)
		}
	}

	return tables, funcMap, nil
}
