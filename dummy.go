package main

import format "fmt"
import "unsafe"

func main() {
	var s *string = new(string)

	*s = "==="
	var sByte []byte = make([]byte, 0)
	format.Println(&sByte)

	sByte = []byte(*s)
	format.Println(*s)
	format.Println(sByte)
	format.Println(&sByte[0])
	pType := unsafe.Pointer(&sByte)
	format.Println(pType)
	var start *int = new(int)
	for *start = 0; *start < len(sByte); *start++ {
		format.Print(string(sByte[*start]))
	}
	var my_int *MyInt = new(MyInt)
	format.Println(*my_int)
}

type MyInt int
