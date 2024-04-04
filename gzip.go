package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	jsutil "gowasmunarchiver/js_util"
	"io"
	"syscall/js"
)

func InflateGZip(contents []byte) ([]byte, error) {
	bytes_reader := bytes.NewReader(contents)
	gzip_reader, err := gzip.NewReader(bytes_reader)
	if err != nil {
		return []byte{}, err
	}
	defer gzip_reader.Close()

	// reaad all the contents
	var deflated bytes.Buffer
	_, err = deflated.ReadFrom(gzip_reader)
	if err != nil {
		return []byte{}, err
	}

	return deflated.Bytes(), nil
}

type DeflateTarFile struct {
	name     string
	contents []byte
}

func InflateTar(contents []byte, callback func(DeflateTarFile)) error {
	bytes_reader := bytes.NewReader(contents)
	tar_reader := tar.NewReader(bytes_reader)

	for {
		header, err := tar_reader.Next()

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		contents := make([]byte, header.Size)
		_, err = tar_reader.Read(contents)
		if err != nil && err != io.EOF {
			return err
		}

		file := DeflateTarFile{name: header.Name, contents: contents}
		callback(file)
	}

	return nil
}

func InflateTarGz(contents []byte, callback func(DeflateTarFile)) error {
	deflated, err := InflateGZip(contents)
	if err != nil {
		return err
	}

	return InflateTar(deflated, callback)
}

func initialiseGZip() {
	gzipObj := js.ValueOf(make(map[string]any))

	gzipObj.Set("gunzip", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return fmt.Sprintf("gunzip: Expected 1 argument, got %d", len(args))
		}
		length, err := jsutil.GetLength(args[0])
		if err != nil {
			return fmt.Sprintf("gunzip: Invalid argument: %s", err)
		}
		contents := make([]byte, length)
		_, err = jsutil.CopyBytesToGo(contents, args[0])
		if err != nil {
			return fmt.Sprintf("gunzip: Could not get data: %s", err)
		}

		deflated, err := InflateGZip(contents)
		if err != nil {
			return fmt.Sprintf("gunzip: Could not deflate gzip: %s", err)
		}

		uint8Array := js.Global().Get("Uint8Array").New(len(deflated))
		_, err = jsutil.CopyBytesToJS(uint8Array, deflated)
		if err != nil {
			return fmt.Sprintf("gunzip: Could not send data: %s", err)
		}

		return uint8Array
	}))

	gzipObj.Set("untargz", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 && len(args) != 2 {
			return fmt.Sprintf("untargz: Expected 1 or 2 arguments, got %d", len(args))
		}
		length, err := jsutil.GetLength(args[0])
		if err != nil {
			return fmt.Sprintf("untargz: Invalid argument: %s", err)
		}
		contents := make([]byte, length)
		_, err = jsutil.CopyBytesToGo(contents, args[0])
		if err != nil {
			return fmt.Sprintf("untargz: Could not get data: %s", err)
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

		err = InflateTarGz(contents, func(file DeflateTarFile) {
			uint8Array := js.Global().Get("Uint8Array").New(len((file.contents)))
			_, err = jsutil.CopyBytesToJS(uint8Array, file.contents)
			if err != nil {
				cbfunc.Invoke(fmt.Sprintf("untargz: Could not send data: %s", err), js.Undefined())
			}
			cbfunc.Invoke(file.name, uint8Array)
		})
		if err != nil {
			return fmt.Sprintf("untargz: Could not deflate tar.gz: %s", err)
		}

		return js.Undefined()
	}))

	js.Global().Set("gzip", gzipObj)
}

func cleanupGZip() {
	js.Global().Delete("gzip")
}
