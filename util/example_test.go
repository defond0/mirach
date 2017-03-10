package util

import "fmt"

func ExampleSplitAt() {
	b, _ := SplitAt([]byte("abc ⌘ efg"), 5)
	fmt.Println(b[1])
	// Output: [140 152 32 101 102]
}

func ExampleSplitStringAt() {
	s, _ := SplitStringAt("abc ⌘ efg", 5)
	fmt.Println(s[1] + ", " + s[2])
	// Output: ⌘ e, fg
}
