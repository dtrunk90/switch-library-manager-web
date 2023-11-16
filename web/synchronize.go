package web

import (
	"encoding/json"
	"net/http"
)

func (web *Web) Synchronize() {
	web.state.IsSynchronizing = true

	go func () {
		web.state.switchDB = nil

		web.updateDB()

		if _, err := web.buildLocalDB(web.localDbManager, true); err != nil {
			web.sugarLogger.Error(err)
		}

		web.state.IsSynchronizing = false
	}()
}

func (web *Web) HandleSynchronize() {
	web.router.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		web.Synchronize()
	}).Methods("POST")

	web.router.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jsonEncoder := json.NewEncoder(w)
		jsonEncoder.Encode(web.state.IsSynchronizing)
	}).Methods("GET")
}
