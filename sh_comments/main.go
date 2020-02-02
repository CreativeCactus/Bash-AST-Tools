package main

import (
	"strings"
	"fmt"
	"mvdan.cc/sh/syntax"
	"bufio"
	"os"
)

type conf struct {
	Depth int // How deep should we recurse?
	Root bool // Should we print the root node(s)?
	Prefix string // Should we indent lines with a prefix?
	Extended bool // Should we show comments?
	Compact bool // Should we print descriptions?
}
func (c *conf) Add (a string) {
	if len(a)<1 {
		return
	}
	switch a[1] {
	case 'h':
		help()
		os.Exit(0)
	case '0':
		c.Depth = 0
	case '1':
		c.Depth = 1
	case '2':
		c.Depth = 2
	case '3':
		c.Depth = 3
	case 'x':
		c.Extended = true
	case 'X':
		c.Extended = false
	case 'c':
		c.Compact = true
	case 'C':
		c.Compact = false
	case 'r':
		c.Root = true
	case 'R':
		c.Root = false
	case 'p':
		c.Prefix = a[2:]
	default:
		stderrf("Unknwon option %s\n", a)
	}
}

// TODO Glob name matching on query eg "run* " returns all subs of run and runtime

// shcom file "a b " # prints functions and comments under the b function inside the a function
// shcom file "a b"  # suggests function names prints functions and comments under the b function inside the a function
// shcom -3 file -c -p"AA" # prints 3 levels deep from the root in compact mode, indenting with AA
const verbose = false
func main(){
	config := conf{ // Default config
		Depth: 0,
		Prefix: "  ",
		Root: true,
		Extended: false,
		Compact: false,
	}

	args := os.Args[1:]
	opts := []string{}
	for _, v := range args {
		if len(v)>0 && v[0]=='-' {
			config.Add(v)
			continue
		}
		opts = append(opts, v)
	}
	if len(opts)<1 || len(opts)>2 {
		help()
		return
	}
	ptr, err := os.Open(opts[0])
	if err != nil {
		stderrf("File reading error: %s", err.Error())
		return
	}

	f := bufio.NewReader(ptr)
	sp := syntax.NewParser(syntax.KeepComments, syntax.Variant(syntax.LangBash))
	sh, err := sp.Parse(f, "")
	if err != nil {
		stderrf("Parse error: %s", err.Error())
		return
	}

	result := Root(sh, opts[0])
	piv("%#v\n", result)

	query := []string{}
	for _, v := range opts[1:] {
		query = append(query, strings.Split(v, " ")...)
	}

	piv("%#v\n", query)
	piv("%#v\n", config)
	if config.Root {
		result.QueryPath(query).Print(config.Depth, config)
		return
	}
	result.QueryPath(query).PrintUnder(config.Depth, config)
}
type FuncGroup []FuncScope
type FuncScope struct {
	Root *syntax.Node // FuncDecl
	Node *syntax.Node // FuncDecl
	Depth int
	Name string
	Desc string // A comment, if any, from the definition(){} line
	Comments []string
	Line uint
	Nested []FuncScope
}
// Query returns the subset of this FG which match the given partial or exact query
func (fg FuncGroup) Query(part string, exact bool) (res FuncGroup) {
	for _, v := range fg {
		if v.Match(part, exact) {
			res = append(res, v)
		}
	}
	return
}
// Match indicates whether a partial or exact name part matches this FS
func (fs FuncScope) Match(name string, exact bool) bool {
	if name == "*" {
		return true
	}
	if exact {
		return fs.Name == name
	}
	return strings.HasPrefix(fs.Name, name)
}
// Query returns the FuncGroup of FS under this FS which match a partial or exact name part
func (fs FuncScope) Query(part string, exact bool) (res FuncGroup) {
	for _, v := range fs.Nested {
		if v.Match(part, exact) {
			res = append(res, v)
		}
	}
	return
}
// QueryPath returns the FuncGroup of FS which are deeply nested according to the path query
func (fg FuncGroup) QueryPath(path []string) (res FuncGroup) {
	if len(path) == 0 {
		return fg
	}
	if len(path) == 1 {
		piv("Partialg: %s from %d\n", path[0], len(fg))
		return fg.Query(path[0], false)
	}
	// There must be more parts to follow, which means that we only want exact matches
	piv("Exactg: %s\n", path[0])
	res = fg.Query(path[0], true)
	return res.QueryPath(path[1:])
}
// QueryPath returns the FuncGroup of FS which are deeply nested according to the path query
func (fs FuncScope) QueryPath(path []string) (res FuncGroup) {
	if len(path) == 0 {
		return FuncGroup{ fs }
	}
	if len(path) == 1 {
		piv("Partial: %s\n", path[0])
		return fs.Query(path[0], false)
	}
	// There must be more parts to follow, which means that we only want exact matches
	res = fs.Query(path[0], true)
	piv("Exact: %s\n", path[0])
	return res.Collect().QueryPath(path[1:])
}
// Collect returns the sum of all nested fs from each fs in the fg
func (fg FuncGroup) Collect() (res FuncGroup) {
	for _, v := range fg {
		res = append(res, v.Nested...)
	}
	return
}

// PrintUnder shows the content of the FSs but not their own root nodes
func (fg FuncGroup) PrintUnder(nested int, c conf) {
	for _, v := range fg {
		v.PrintUnder(nested, c)
	}
}
// Print calls Print on each underlying FS of this FG
func (fg FuncGroup) Print(nested int, c conf) {
	for _, v := range fg {
		v.Print(nested, c)
	}
}
// PrintUnder shows the content of a FS but not the current node
func (fs FuncScope) PrintUnder(nested int, c conf) {
	FuncGroup(fs.Nested).Print(nested, c)
}
// Print shows the content of a FS with an optional recursion depth, extended comments and indent (prefix) string
func (fs FuncScope) Print(nested int, c conf) {
	prefix := fmt.Sprintf("%s%s ", strings.Repeat(c.Prefix, fs.Depth), fs.Name)
	fmt.Printf("%s", prefix)
	if !c.Compact {
		fmt.Printf("%s", fs.Desc)
	}
	fmt.Printf("\n")
	if c.Extended {
		for _, w := range fs.Comments {
			fmt.Printf("%s%s\n", strings.Repeat(" ", len(prefix)), w)
		}
	}
	if nested==0 {
		return
	}
	for _, v := range fs.Nested {
		v.Print(nested-1, c)
	}
}
func Root(root syntax.Node, name string) (fs FuncScope) {
	fs = Walk(root, root, 0)
	fs.Name = name
	return
}
func Walk(root, node syntax.Node, level int) (fs FuncScope) {
	firstNode := true

	fs.Root = &root
	fs.Node = &node
	fs.Depth = level
	fs.Line = node.Pos().Line()

	syntax.Walk(node, func (node syntax.Node) bool {
		if firstNode {
			firstNode = false
			return true
		}
		switch n := node.(type) {
		case *syntax.FuncDecl:
			nest := Walk(root, n, level+1)
			nest.Name = n.Name.Value
			fs.Nested = append(fs.Nested, nest)
			return false
		case *syntax.Comment:
			if n.Text[0]=='#' {
				if n.Pos().Line() == fs.Line {
					fs.Desc = n.Text
				} else {
					fs.Comments = append(fs.Comments, n.Text)
				}
			}
		}
		return true
	})
	return
}
var version = `0.0.1`
var commitHash string
func help(){
	stderrfln(`shcom: CC/bash_ast/sh_comments v%s-%s`, version, commitHash)
	stderrfln(` usage: %s file "query" `, os.Args[0])
	stderrfln(` Reads file, scopes to a query if provided`)
	stderrfln(` Prints function names and ## comments matching the query in file`)
	stderrfln(` Query is a space-delimited list of nested functions in file`)
	stderrfln(` If query ends in a space then functions and comments under the query are listed`)
	stderrfln(` Otherwise functions matching the partial query will be shown instead`)
	stderrfln(` Nothing is printed if the query does not match any function in the file`)
}
func stderrfln(s string, arg ...interface{}) {
	stderrf(fmt.Sprintf("%s\n",s), arg...)
}
func stderrf(s string, arg ...interface{}) {
	fmt.Fprintf(os.Stderr, s, arg...)
}
// Print if verbose
func piv(s string, arg ...interface{}) {
	if !verbose {
		return
	}
	fmt.Printf(s, arg...)
}

