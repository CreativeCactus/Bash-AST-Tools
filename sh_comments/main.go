package main

import (
	"strings"
	"fmt"
	"mvdan.cc/sh/syntax"
	"bufio"
	"os"
)

type conf struct {
	Depth int
	Prefix string
	Extended bool
}
func (c *conf) Add (a string) {
	if len(a)<1 {
		return
	}
	switch a[1] {
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
	case 'c':
		c.Extended = false
	case 'p':
		c.Prefix = a[2:]
	default:
		stderrf("Unknwon option %s\n", a)
	}
}


// shcom file "a b " # prints functions and comments under the b function inside the a function
// shcom file "a b"  # suggests function names prints functions and comments under the b function inside the a function
// shcom -3 file -c -p"AA" # prints 3 levels deep from the root in compact mode, indenting with AA
func main(){
	config := conf{ // Default config
		Depth: 1,
		Prefix: "  ",
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
	//fmt.Printf("%#v\n", result)

	query := []string{}
	for _, v := range opts[1:] {
		query = append(query, strings.Split(v, " ")...)
	}

	//fmt.Printf("%#v\n", query)
	result.QueryPath(query).Print(config.Depth, config.Extended, config.Prefix)



/*
	results := []FunctionDescriptor{}
	syntax.Walk(sh, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Comment:
			if n.Text[0]=='#' {
				fmt.Printf(" %s\n", n.Text)
			}
		case *syntax.FuncDecl:
			results = append(results, getFunction(n))
			// fmt.Printf("%s", n.Name.Value)
			return false
		default:
			//fmt.Printf("%+v", n)
			break
		}
		return true // traverse down into
	})

*/
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
		if !v.Match(part, exact) {
			continue
		}
		res = append(res, v)
	}
	return
}
// QueryPath returns the FuncGroup of FS which are deeply nested according to the path query
func (fs FuncScope) QueryPath(path []string) (res FuncGroup) {
	if len(path) == 0 {
		return FuncGroup{ fs }
	}
	res = fs.Query(path[0], false)
	if len(path) == 1 {
		return
	}
	// There must be more parts to follow, which means that we only want exact matches
	return res.Query(path[0], true)
}
func (fg FuncGroup) Print(nested int, extended bool, indent string) {
	for _, v := range fg {
		v.Print(nested, extended, indent)
	}
}
// PrintUnder shows the content of a FS but not the current node
func (fs FuncScope) PrintUnder(nested int, extended bool, indent string) {
	FuncGroup(fs.Nested).Print(nested, extended, indent)
}
// Print shows the content of a FS with an optional recursion depth, extended comments and indent string
func (fs FuncScope) Print(nested int, extended bool, indent string) {
	prefix := fmt.Sprintf("%s%s ", strings.Repeat(indent, fs.Depth), fs.Name)
	fmt.Printf("%s%s\n", prefix, fs.Desc)
	if extended {
		for _, w := range fs.Comments {
			fmt.Printf("%s%s\n", strings.Repeat(" ", len(prefix)), w)
		}
	}
	if nested==0 {
		return
	}
	for _, v := range fs.Nested {
		v.Print(nested-1, extended, indent)
	}
}
/*
		fmt.Printf("%s\t%s\n",v.Name,v.Desc)
		if extended {
			for _, w := range v.Comments {
				fmt.Printf("\t%s\n",w)
			}
		}
		if nested==0 {
			continue
		}
		for _, n := range v.Nested {
			n.Print(nested-1, extended)
		}


// Replace with direct call to in/on
func Describe(root syntax.Node, path []string) (fs FuncScope, err error) {
	breakWalk := false
	firstNode := true
	notFirst := func(){
		firstNode = false
	}
	if len(path) > 0 {
		err, fs = WalkIn(root, path)
		if err != nil {
			fmt.Println("Error bad 79")
		}
		return err, fd
	}
	err, df = WalkOn(root)
	if err != nil {
		fmt.Println("Error bad 85")
	}
	return err, df
}
*/
/*
func WalkOn(root *syntax.Node)  (fs FuncScope) {
	firstNode := true
	syntax.Walk(root, func(node syntax.Node) bool {
		if first {
			first = false
			return true
		}
		switch n := node.(type) {
		case *syntax.FuncDecl:
			fs.Nested = WalkOne{ Name: n.Name.Value }
			return false
		case *syntax.Comment:
			if n.Text[0]=='#' {
				if n.Pos().Line() == fd.Line {
					fd.Desc = n.Text
				} else {
					fs.Comments = append(fs.Comments, n.Text)
				}
			}
		}
	})
}
*/
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
	// Could set a default name, desc here
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
/*
func getFunction(n *syntax.FuncDecl, path []string) (fd FunctionDescriptor) {
	breakWalk := false
	firstNode := true
	fd.Root = n
	fd.Name = n.Name.Value
	fd.Line = n.Pos().Line()

	syntax.Walk(n, func(node syntax.Node) bool {
		if breakWalk {
			return false
		}
		defer func(){
			firstNode = false
		}()
		switch n := node.(type) {
		case *syntax.FuncDecl:
			if firstNode {
				return true
			}
			if len(path) > 0 && path[0] == n.Name.Value {
				fd = getFunction(n, path[1:])
				breakWalk = true
				return true
			}
			return false
		case *syntax.Comment:
			if n.Text[0]=='#' {
				if n.Pos().Line() == fd.Line {
					fd.Desc = n.Text
				} else {
	                                fd.Comments = append(fd.Comments, n.Text)
				}
                        }
		}
		//fmt.Println("node")
		return true
	})
	return fd
}
*/
var version = `0.0.1`
var commitHash string
func help(){
	stderrf(`shcom: CC/bash_ast/sh_comments v%s-%s\n`, version, commitHash)
	stderrf(` usage: %s file "query" \n`, os.Args[0])
	stderrf(` Reads file, scopes to a query if provided\n`)
	stderrf(` Prints function names and ## comments matching the query in file\n`)
	stderrf(` Query is a space-delimited list of nested functions in file\n`)
	stderrf(` If query ends in a space then functions and comments under the query are listed\n`)
	stderrf(` Otherwise functions matching the partial query will be shown instead\n`)
	stderrf(` Nothing is printed if the query does not match any function in the file\n`)
}
func stderrf(s string, arg ...interface{}) {
	fmt.Fprintf(os.Stderr, s, arg...)
}
