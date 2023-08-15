package web

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

func (web *Web) HandleResources() {
	fSys, err := fs.Sub(web.embedFS, "resources/static")
	if err != nil {
		web.sugarLogger.Error(fmt.Errorf("getting static files failed: %w", err))
		log.Fatal(err)
	}
	http.Handle("/resources/static/", http.StripPrefix("/resources/static/", http.FileServer(http.FS(fSys))))

	fSys, err = fs.Sub(web.embedFS, "node_modules")
	if err != nil {
		web.sugarLogger.Error(fmt.Errorf("getting vendor files failed: %w", err))
		log.Fatal(err)
	}
	http.Handle("/resources/vendor/", http.StripPrefix("/resources/vendor/", http.FileServer(http.FS(fSys))))
}
