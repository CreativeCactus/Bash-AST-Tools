package main

import (
	"fmt"
//	"io/ioutil"
	"mvdan.cc/sh/syntax"
	"bufio"
	"os"
)


func main(){
	args := os.Args[1:]
	if len(args)!=1 {
		help()
		return
	}
	ptr, err := os.Open(args[0])
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	f := bufio.NewReader(ptr)

	sp := syntax.NewParser(syntax.KeepComments, syntax.Variant(syntax.LangBash))
	sh, err := sp.Parse(f, "")
	if err != nil {
		return // err
	}

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
	for _, v := range results {
		fmt.Printf("%s\t%s\n",v.Name,v.Desc)
		for _, w := range v.Comments {
			fmt.Printf("\t%s\n",w)
		}
	}
}
type FunctionDescriptor struct {
	Name string
	Desc string // A comment, if any, from the definition(){} line
	Comments []string
}
func getFunction(n *syntax.FuncDecl) (fd FunctionDescriptor) {
	first := true
	fd.Name = n.Name.Value
	line := n.Pos().Line()
	syntax.Walk(n, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.FuncDecl:
			isFirst := first
			first = false
			return isFirst
		case *syntax.Comment:
			if n.Text[0]=='#' {
				if n.Pos().Line() == line {
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
func help(){
	fmt.Println(`./bash_ast file`)
}
