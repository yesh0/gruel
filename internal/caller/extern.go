// This package is separated from grueljit
// in that Go forbids using CGO and Plan 9 assembly in the same package.
package caller

//go:generate go run ../../build/grueljit/caller.go -out caller.s -stubs caller.go
