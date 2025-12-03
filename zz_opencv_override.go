//go:build windows
// +build windows

package main

/*
#cgo CFLAGS: -IC:/opencv/build/include
#cgo LDFLAGS: -LC:/opencv/build/x64/vc16/lib -lopencv_world4120
*/
import "C"
