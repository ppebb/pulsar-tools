//go:build windows

package main

import _ "embed"

//go:embed resources/wbmgt.exe
var wbmgtBytes []byte
