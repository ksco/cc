package tests

import (
	"cc/cc"
	"fmt"
	"github.com/bytecodealliance/wasmtime-go"
	"reflect"
	"strings"
	"testing"
)

type Assert struct {
	t *testing.T
}

func (a Assert) Eval(expected interface{}, s string) {
	sb := new(strings.Builder)
	err := cc.Compile(sb, []rune(s))
	if err != nil {
		a.t.Errorf("Compile failed, error:\n%s\ncode: %s", err.Error(), s)
	}

	wasm, err := wasmtime.Wat2Wasm(sb.String())
	if err != nil {
		a.t.Errorf("Wat2Wasm failed, error:\n%s\ncode: %s", err.Error(), s)
	}
	store := wasmtime.NewStore(wasmtime.NewEngine())
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		a.t.Errorf("Create Wasm module failed, error:\n%s\ncode: %s", err.Error(), s)
	}
	instance, err := wasmtime.NewInstance(store, module, nil)
	if err != nil {
		a.t.Errorf("Create Wasm instance failed, error:\n%s\ncode: %s", err.Error(), s)
	}
	run := instance.GetExport(store, "main").Func()
	result, err := run.Call(store)
	if err != nil {
		a.t.Errorf("Run Wasm instance failed, error:\n%s\ncode: %s", err.Error(), s)
	}
	if result != expected {
		fmt.Printf("%v, %v\n", reflect.TypeOf(result), reflect.TypeOf(expected))
		a.t.Errorf("Result error, expected: %v, got %d, code: %s", expected, result, s)
		return
	}
}
