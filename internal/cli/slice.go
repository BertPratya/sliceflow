package cli

import "fmt"

// runSliceCreate runs "sliceflow slice create". args are the words after
// "create". For now it only shows that the routing reached it.
func runSliceCreate(args []string) int {
	fmt.Println("Create called")
	return 0
}
