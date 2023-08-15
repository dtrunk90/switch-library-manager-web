package web

import (
	"encoding/json"
	"fmt"
	"github.com/ggicci/httpin"
	"github.com/justinas/alice"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type ErrorResponse struct {
	GlobalError GlobalError  `json:"globalError"`
	FieldErrors []FieldError `json:"fieldErrors"`
}

type GlobalError struct {
	StrongMessage string `json:"strongMessage"`
	Message       string `json:"message"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	StrongMessage string `json:"strongMessage"`
	Message       string `json:"message"`
}

type FilteredPageData func(filter *TitleItemFilter) any
type Validate func(value any) ErrorResponse
type OnSuccess func(value any) SuccessResponse
type PageData func() any

func hasErrors(errorResponse ErrorResponse) bool {
	return errorResponse.GlobalError != GlobalError{} || len(errorResponse.FieldErrors) > 0
}

func (web *Web) HandleFiltered(pattern string, filteredPageData FilteredPageData, fs fs.FS, fsPatterns ...string) {
	tmpl := template.New("layout").Funcs(funcMap)
	tmpl, err := tmpl.ParseFS(fs, fsPatterns...)

	if err != nil {
		web.sugarLogger.Error(fmt.Errorf("parsing template failed: %w", err))
		log.Fatal(err)
	}

	http.Handle(pattern, alice.New(httpin.NewInput(TitleItemFilter{})).ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := r.Context().Value(httpin.Input).(*TitleItemFilter)
		if err := tmpl.ExecuteTemplate(w, "layout", filteredPageData(filter)); err != nil {
			web.sugarLogger.Error(fmt.Errorf("executing template failed: %w", err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
}

func (web *Web) HandleValidated(pattern string, inputStruct interface{}, pageData PageData, validate Validate, onSuccess OnSuccess, fs fs.FS, fsPatterns ...string) {
	tmpl := template.New("layout").Funcs(funcMap)
	tmpl, err := tmpl.ParseFS(fs, fsPatterns...)

	if err != nil {
		web.sugarLogger.Error(fmt.Errorf("parsing template failed: %w", err))
		log.Fatal(err)
	}

	http.Handle(pattern, alice.New(httpin.NewInput(inputStruct)).ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
			case "GET":
				if err := tmpl.ExecuteTemplate(w, "layout", pageData()); err != nil {
					web.sugarLogger.Error(fmt.Errorf("executing template failed: %w", err))
					w.WriteHeader(http.StatusInternalServerError)
				}
			case "POST":
				value := r.Context().Value(httpin.Input)
				jsonEncoder := json.NewEncoder(w)

				if errorResponse := validate(value); hasErrors(errorResponse) {
					w.WriteHeader(http.StatusBadRequest)
					jsonEncoder.Encode(errorResponse)
					return
				}

				jsonEncoder.Encode(onSuccess(value))
			default:
				web.sugarLogger.Error(fmt.Errorf("Unsupported method: %s", r.Method))
				w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func (web *Web) Handle(pattern string, pageData PageData, fs fs.FS, fsPatterns ...string) {
	tmpl := template.New("layout").Funcs(funcMap)
	tmpl, err := tmpl.ParseFS(fs, fsPatterns...)

	if err != nil {
		web.sugarLogger.Error(fmt.Errorf("parsing template failed: %w", err))
		log.Fatal(err)
	}

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.ExecuteTemplate(w, "layout", pageData()); err != nil {
			web.sugarLogger.Error(fmt.Errorf("executing template failed: %w", err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
