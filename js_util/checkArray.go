package jsutil

import "syscall/js"

func CheckIsArray(v js.Value) (res bool) {
	defer func() {
		if r := recover(); r != nil {
			res = false
		}
	}()

	res = js.Global().Get("Array").Call("isArray", v).Truthy()
	return
}

func CheckIsUint8Array(v js.Value) (res bool) {
	defer func() {
		if r := recover(); r != nil {
			res = false
		}
	}()

	res = v.Get("constructor").Get("name").String() == "Uint8Array"
	return
}
