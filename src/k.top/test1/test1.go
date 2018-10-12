package main

/*
#include <stdio.h>
	void hello() {
        printf("Hello, Cgo! -- From C world.\n");
    }
*/
import "C"
import "fmt"

func Hello() int {
	C.hello()
	return 1
}

func main() {
	hello := Hello()
	fmt.Printf("%v", hello)
}
