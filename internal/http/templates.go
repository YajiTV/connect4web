package httphandler

import "io/fs"

var templateFS fs.FS

func SetTemplateFS(fsys fs.FS) { templateFS = fsys }
