package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	jsutil "gowasmunarchiver/js_util"
	"syscall/js"
)

func DeflateZip(jsContents js.Value, fileCallback js.Value) error {
	length, err := jsutil.GetLength(jsContents)
	if err != nil {
		return fmt.Errorf("DeflateZip: Could not create destination byte[]: %v", err)
	}

	dst := make([]byte, length)

	jsutil.CopyBytesToGo(dst, jsContents)
	bytesReader := bytes.NewReader(dst)
	reader, err := zip.NewReader(bytesReader, int64(length))
	if err != nil {
		return fmt.Errorf("DeflateZip: Could not create zip reader: %v", err)
	}
	for _, file := range reader.File {
		r, err := file.Open()
		defer r.Close()
		if err != nil {
			fileCallback.Invoke(fmt.Sprintf("Could not load file %s: %v", file.Name, err), js.Undefined())
		}
		data := make([]byte, file.UncompressedSize64)
		r.Read(data)
		uint8Array := js.Global().Get("Uint8Array").New(len(data))
		js.CopyBytesToJS(uint8Array, data)
		fileCallback.Invoke(file.Name, uint8Array)
	}

	return nil
}

func initialiseZip() {
	zipObj := js.ValueOf(make(map[string]any))

	zipObj.Set("deflateZip", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 && len(args) != 2 {
			return fmt.Sprintf("deflateZip: Expected 1 or 2 arguments, got %d", len(args))
		}
		var cbfunc js.Value
		if len(args) == 1 {
			cbfunc = js.Global().Get("console").Get("log")
		} else {
			cbfunc = args[1]
		}
		if cbfunc.Type() != js.TypeFunction {
			cbfunc = js.Global().Get("console").Get("log")
		}
		err := DeflateZip(args[0], cbfunc)
		if err != nil {
			return fmt.Sprintf("deflateZip: %v", err)
		}
		return js.Undefined()
	}))

	js.Global().Set("zip", zipObj)
}

func cleanupZip() {
	js.Global().Delete("zip")
}
