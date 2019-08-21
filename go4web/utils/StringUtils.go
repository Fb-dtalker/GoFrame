package utils

import "unsafe"

func BytesToString(b []byte) (s string) {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) (b []byte) {
	return *(*[]byte)(unsafe.Pointer(&s))
}
