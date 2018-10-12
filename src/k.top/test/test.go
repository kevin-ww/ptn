package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func main() {
	fmt.Printf("haha")

	shim.Success([]byte(`haha`))
}
