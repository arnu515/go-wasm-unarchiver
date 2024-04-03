package jsutil

import "syscall/js"
import "fmt"

func GetLength(src js.Value) (length int, err error) {
	defer func() {
		if r := recover(); r != nil {
			length = 0
			err = fmt.Errorf("GetLength: %v", r)
		}
	}()

	length = src.Length()
	err = nil

	return
}
