package httphandler

import "io/fs"

// templateFS holds the embedded templates filesystem
var templateFS fs.FS

// SetTemplateFS injects the templates filesystem at startup
func SetTemplateFS(fsys fs.FS) { templateFS = fsys }
