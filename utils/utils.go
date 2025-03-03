package utils

import "container/list"

func ListToStringArray(l *list.List) []string {
	// converts a linkedlist into a string array
	arr := make([]string, l.Len())
	node := l.Front()
	slider := 0
	for node != nil {
		arr[slider] = node.Value.(string) // set array val
		slider++                          // increment slider
		node = node.Next()                // move to the next node
	}
	return arr // return the array
}
