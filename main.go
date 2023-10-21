//go:build js && wasm

package main

import (
	"fmt"
	"io"
	"net/http"
	"syscall/js"
	"time"
)

var (
	counter int
)

func main() {
	// Register a callback function
	js.Global().Set("hellowasm", js.FuncOf(hello))
	js.Global().Set("GoWebRequestFunc", GoWebRequestFunc())
	// Update the DOM content from WASM
	updateDOMContent()
	go updateTime()
	// Wait for ever
	select {}
}

func hello(this js.Value, p []js.Value) interface{} {
	counter++
	return js.ValueOf(fmt.Sprintf("Hello from Go WASM! Counter is: %d", counter))
}

// Simple update of content
func updateDOMContent() {
	document := js.Global().Get("document")
	element := document.Call("getElementById", "myParagraph")
	element.Set("innerText", "This line was updated directly from Go WASM after load!")
}

// Continuous updating content on the page
func updateTime() {
	for {
		document := js.Global().Get("document")
		element := document.Call("getElementById", "time")
		element.Set("innerText", time.Now().Format(time.RFC3339))
		time.Sleep(time.Second)
	}
}

func GoWebRequestFunc() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Get the URL as argument
		requestUrl := args[0].String()
		// return a Promise because HTTP requests are blocking in Go
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]
			go func() {
				// The HTTP request
				res, err := http.DefaultClient.Get(requestUrl)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}
				defer res.Body.Close()

				// Read the response body
				data, err := io.ReadAll(res.Body)
				if err != nil {
					// Handle errors here too
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}
				arrayConstructor := js.Global().Get("Uint8Array")
				dataJS := arrayConstructor.New(len(data))
				js.CopyBytesToJS(dataJS, data)
				responseConstructor := js.Global().Get("Response")
				response := responseConstructor.New(dataJS)
				resolve.Invoke(response)
			}()
			return nil
		})
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}
