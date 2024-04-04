package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	jsutil "gowasmunarchiver/js_util"
	"syscall/js"
)

func DeflateZip(contents []byte, callback func(name string, contents []byte)) error {
	bytesReader := bytes.NewReader(contents)
	reader, err := zip.NewReader(bytesReader, int64(len(contents)))
	if err != nil {
		return fmt.Errorf("DeflateZip: Could not create zip reader: %v", err)
	}
	for _, file := range reader.File {
		r, err := file.Open()
		defer r.Close()
		if err != nil {
			callback(fmt.Sprintf("Could not load file %s: %v", file.Name, err), []byte{})
		}
		data := make([]byte, file.UncompressedSize64)
		r.Read(data)
		callback(file.Name, data)
	}

	return nil
}

func CreateZip(files map[string][]byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	for name, contents := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return nil, fmt.Errorf("CreateZip: Could not create zip writer: %v", err)
		}
		_, err = writer.Write(contents)
		if err != nil {
			return nil, fmt.Errorf("CreateZip: Could not write to zip writer: %v", err)
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("CreateZip: Could not close zip writer: %v\n", err)
	}
	return buf.Bytes(), nil
}

func initialiseZip() {
	zipObj := js.ValueOf(make(map[string]any))

	zipObj.Set("deflateZip", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 && len(args) != 2 {
			return fmt.Sprintf("deflateZip: Expected 1 or 2 arguments, got %d", len(args))
		}

		if !jsutil.CheckIsUint8Array(args[0]) {
			return fmt.Sprintf("deflateZip: Expected Uint8Array, got %s", args[0].Get("constructor").Get("name").String())
		}
		length, err := jsutil.GetLength(args[0])
		if err != nil {
			return fmt.Errorf("deflateZip: Could not get length of Uint8Array: %v", err)
		}
		contents := make([]byte, length)
		_, err = jsutil.CopyBytesToGo(contents, args[0])
		if err != nil {
			return fmt.Errorf("deflateZip: Could not create destination byte[]: %v", err)
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

		err = DeflateZip(contents, func(name string, contents []byte) {
			if len(contents) == 0 {
				cbfunc.Invoke(name, js.Undefined())
			} else {
				uint8Array := js.Global().Get("Uint8Array").New(len(contents))
				js.CopyBytesToJS(uint8Array, contents)
				cbfunc.Invoke(name, uint8Array)
			}
		})
		if err != nil {
			return fmt.Sprintf("deflateZip: %v", err)
		}
		return js.Undefined()
	}))

	zipObj.Set("createZip", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return fmt.Sprintf("createZip: Expected 1 argument, got %d", len(args))
		}

		if !jsutil.CheckIsArray(args[0]) {
			return fmt.Sprintf("createZip: Expected Array, got %s", args[0].Get("constructor").Get("name").String())
		}

		files := make(map[string][]byte)
		js.Global().Get("Array").Get("prototype").Get("forEach").Call("call", args[0], js.FuncOf(func(this js.Value, args []js.Value) any {
			nameJs := args[0].Get("name")
			if nameJs.IsUndefined() || nameJs.IsNull() {
				return fmt.Sprintf("createZip: Index %s of array: Expected name to be defined, got %s", args[1], nameJs.String())
			}
			contentsJs := args[0].Get("contents")
			if !jsutil.CheckIsUint8Array(contentsJs) {
				return fmt.Sprintf("createZip: Index %s of array: Expected contents to be Uint8Array, got %s", args[1], contentsJs.Get("constructor").Get("name").String())
			}

			length, err := jsutil.GetLength(contentsJs)
			if err != nil {
				return fmt.Errorf("createZip: Index %s of array: Could not get length of contents: %v", args[1], err)
			}
			contents := make([]byte, length)
			_, err = jsutil.CopyBytesToGo(contents, contentsJs)
			if err != nil {
				return fmt.Errorf("createZip: Index %s of array: Could not create destination byte[]: %v", args[1], err)
			}

			files[nameJs.String()] = contents
			return js.Undefined()
		}))

		zip, err := CreateZip(files)
		if err != nil {
			return fmt.Sprintf("createZip: Could not create zip: %v", err)
		}

		uint8Array := js.Global().Get("Uint8Array").New(len(zip))
		_, err = jsutil.CopyBytesToJS(uint8Array, zip)
		if err != nil {
			return fmt.Errorf("createZip: Could not copy zip to Uint8Array: %v", err)
		}

		return uint8Array
	}))

	js.Global().Set("zip", zipObj)
}

func cleanupZip() {
	js.Global().Delete("zip")
}
