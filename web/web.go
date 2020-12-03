package web

import (
	"encoding/json"
	"fmt"
	"github.com/sparrowhawktech/go-lib/util"
	"net/http"
	"reflect"
	"strings"
)

type HttpError struct {
	StatusCode int
	Error      interface{}
}

func ParseParamOrBody(r *http.Request, o interface{}) {
	s := r.URL.Query().Get("body")
	if len(s) > 0 {
		util.CheckErr(json.NewDecoder(strings.NewReader(s)).Decode(o))
	} else {
		util.CheckErr(json.NewDecoder(r.Body).Decode(o))
	}
}

func InterceptCORS(delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			header := r.Header.Get("Access-Control-Request-Headers")
			if len(header) > 0 {
				w.Header().Add("Access-Control-Allow-Headers", header)
			}
		} else {
			delegate(w, r)
		}
	}
}

func InterceptFatal(delegate func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer catchFatal(w, r)
		delegate(w, r)
	}
}

func catchFatal(writer http.ResponseWriter, r *http.Request) {
	if e := recover(); e != nil {
		util.Logger("error").Printf("Error executing %s", r.URL.String())
		errorType := reflect.TypeOf(e)
		if errorType.Kind() == reflect.Struct {
			jsonBytes := util.Marshal(e)
			util.ProcessPanic(string(jsonBytes))
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write(jsonBytes)
		} else {
			util.ProcessPanic(e)
			http.Error(writer, fmt.Sprint(e), http.StatusInternalServerError)
		}
	}
}

func JsonResponse(i interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "Application/json")
	util.JsonEncode(i, w)
}
