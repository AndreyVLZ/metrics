package hash_test

import (
	"encoding/hex"
	"fmt"

	"github.com/AndreyVLZ/metrics/pkg/hash"
)

func ExampleSHA256() {
	key := []byte("secret key")
	data := []byte("data from hash")
	resData, _ := hash.SHA256(data, key)

	fmt.Println(hex.EncodeToString(resData))

	// Output:
	// 3f5b3476c2533cb11633687ed8e3599cede1a1751b2c8531d19781180dc5b0cc
}

func ExampleValidMAC() {
	key := []byte("secret key")
	data := []byte("data from hash")
	mac := "3f5b3476c2533cb11633687ed8e3599cede1a1751b2c8531d19781180dc5b0cc"

	isEqual, _ := hash.ValidMAC(mac, data, key)
	fmt.Println(isEqual)

	// Output:
	// true
}
