package main

import (
	"bali-rent/pkg/files"
)

func main() {
	// os.Setenv("PORT", "1234")
	// os.Setenv("FILES_DIR", "../../images")
	files.RunFileServer()
}
