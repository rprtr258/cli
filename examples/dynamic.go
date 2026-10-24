// `dynamic.rs` - you can construct a parser at runtime without having a concrete type too.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type dynamicValueKind uint8

const (
	dynamicValueBool dynamicValueKind = iota
	dynamicValueNumber
	dynamicValueString
)

type dynamicValue struct {
	kind   dynamicValueKind
	b      bool
	number int
	str    string
}

type dynamicItem struct {
	name  string
	value dynamicValue
}

func Dynamic() {
	number := func(name string) cli.Parser[dynamicItem] {
		return cli.Long(name).
			Argument("NUM", cli.ParseInt).
			Parser().
			Map(func(n int) dynamicItem {
				return dynamicItem{name, dynamicValue{kind: dynamicValueNumber, number: n}}
			})
	}
	boolean := func(name string) cli.Parser[dynamicItem] {
		return cli.Long(name).
			Switch().
			Map(func(b bool) dynamicItem {
				return dynamicItem{name, dynamicValue{kind: dynamicValueBool, b: b}}
			})
	}
	str := func(name string) cli.Parser[dynamicItem] {
		return cli.Long(name).
			Help("this can use a help message").
			Argument("NUM", cli.ParseString).
			Parser().
			Map(func(s string) dynamicItem {
				return dynamicItem{name, dynamicValue{kind: dynamicValueString, str: s}}
			})
	}

	type item struct {
		name string
		kind dynamicValueKind
	}
	items := []item{
		{"banana", dynamicValueBool},
		{"width", dynamicValueNumber},
		{"name", dynamicValueString},
	}

	acc := cli.Pure([]dynamicItem{})
	for _, it := range items {
		var cur cli.Parser[dynamicItem]
		switch it.kind {
		case dynamicValueBool:
			cur = boolean(it.name)
		case dynamicValueNumber:
			cur = number(it.name)
		default:
			cur = str(it.name)
		}

		acc = cli.Pair(acc, cur).Parser().Map(
			func(p fun.Pair[[]dynamicItem, dynamicItem]) []dynamicItem {
				return append(p.K, p.V)
			},
		)
	}

	fmt.Println(acc.ToOptions().Run(os.Args))
}
