/**
This function is a wrapper around syscall/js's CopyBytesToGo function.
It returns an error instead of panicking.
*/

package jsutil

import "syscall/js"
import "fmt"

func CopyBytesToGo(dst []byte, src js.Value) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			n = 0
			err = fmt.Errorf("CopyBytesToGo: %v", r)
		}
	}()

	n = js.CopyBytesToGo(dst, src)
	err = nil

	return
}
