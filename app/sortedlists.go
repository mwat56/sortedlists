/*
Copyright © 2024  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"fmt"

	sl "github.com/mwat56/sortedlists"
)

func main() {
	// Example usage
	ints := sl.NewSlice([]int{5, 3, 4, 1, 2}, false)

	ints.Insert(6)
	fmt.Println(ints.Data()) // Output: [1 2 3 4 5 6]

	ints.Delete(3)
	fmt.Println(ints.Data()) // Output: [1 2 4 5 6]

	ints.Rename(4, 7)
	fmt.Println(ints.Data()) // Output: [1 2 5 6 7]
} // main()

/* _EoF_ */
