// Example for goto
package main

import (
	"fmt"
)

func main() {
	goto first
second:
	fmt.Println("SECOND.")
	goto done
first:
	fmt.Println("FIRST.")
	goto second
done:
	fmt.Println("DONE.")
}
