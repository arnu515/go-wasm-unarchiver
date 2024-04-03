/**
This function is a wrapper around syscall/js's CopyBytesToJS function.
It returns an error instead of panicking.
*/

package jsutil

import "syscall/js"
import "fmt"

func CopyBytesToJS(dst js.Value, src []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			n = 0
			err = fmt.Errorf("CopyBytesToGo: %v", r)
		}
	}()

	n = js.CopyBytesToJS(dst, src)
	err = nil

	return
}
