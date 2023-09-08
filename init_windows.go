package goblet

import "mime"

func init() {
	mime.AddExtensionType(".js", "application/javascript")
}
