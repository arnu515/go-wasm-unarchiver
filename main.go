package main

import "syscall/js"

var stopChan chan bool

func main() {
	println("Hello from Go!")

	stopChan = make(chan bool, 1)

	js.Global().Set("stop", js.FuncOf(func(this js.Value, args []js.Value) any {
		stopChan <- true

		return js.Undefined()
	}))

	initialiseZip()

	<-stopChan

	cleanupZip()

	println("Goodbye from Go!")
}
