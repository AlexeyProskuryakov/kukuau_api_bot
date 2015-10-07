package main

import (
	"msngr/taxi/set"
	"fmt"
)

func main() {
	s1 := set.NewSet()
	s2 := set.NewSet()

	s1.Add("foo")
	s1.Add("bar")
	s1.Add("baz")

	s2.Add("baz")
	s2.Add("bak")
	s2.Add("baf")

	fmt.Print(s1.SymmetricDifference(s2))

}