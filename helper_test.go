package trial

import (
	"fmt"
)

func testPointer[T primitives](v T) {
	p := Pointer(v)
	fmt.Printf("%T=%v ", p, *p)
}
func ExamplePointer() {
	testPointer(1)
	testPointer(int8(2))
	testPointer(int16(3))
	testPointer(int32(4))
	testPointer(int64(5))

	fmt.Println("|")
	testPointer(uint(10))
	testPointer(uint8(11))
	testPointer(uint16(12))
	testPointer(uint32(13))
	testPointer(uint64(14))

	fmt.Println("|")
	testPointer("hello")
	testPointer(false)
	testPointer(float32(12.34))
	testPointer(99.99)

	//Output: *int=1 *int8=2 *int16=3 *int32=4 *int64=5 |
	// *uint=10 *uint8=11 *uint16=12 *uint32=13 *uint64=14 |
	// *string=hello *bool=false *float32=12.34 *float64=99.99
}
