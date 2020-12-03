package main

import (
	"flag"
	"github.com/gabrielmorenobrc/go-lib/auth"
	"github.com/gabrielmorenobrc/go-lib/react"
	"github.com/gabrielmorenobrc/go-lib/sql"
	"github.com/gabrielmorenobrc/go-lib/util"
	"github.com/gabrielmorenobrc/go-lib/web"
	"net/http"
	"path/filepath"
)

type Config struct {
	HttpPort       *string              `json:"httpPort"`
	DatabaseConfig *sql.DatabaseConfig  `json:"databaseConfig"`
	LogToConsole   *bool                `json:"logToConsole"`
	LogTags        []string             `json:"logTags"`
	SessionConfig  *auth.SessionsConfig `json:"sessionsConfig"`
}

var conf = flag.String("conf", "conf.json", "Config")

func main() {
	flag.Parse()
	var config Config
	util.LoadConfig(*conf, &config)

	util.ConfigLoggers("container.log", 2000000, 10, *config.LogToConsole, config.LogTags...)

	handleUi("admin", "/admin/")

	sessionDataProvider := &auth.MockDataProvider{}
	sessionManager := auth.NewSessionManager(sessionDataProvider, *config.SessionConfig)
	sessionManager.Shrink()
	sessionManager.Load()

	util.CheckErr(http.ListenAndServe("0.0.0.0:"+*config.HttpPort, nil))

}

func handleUi(name string, path string) {
	isdev := !util.FileExists("ui/" + name)
	var folder string
	if isdev {
		folder = "../" + name + "/build"
	} else {
		folder = "ui/" + name
	}
	abs, _ := filepath.Abs(folder)
	println(abs)
	fs := http.FileServer(http.Dir(abs))
	sp := http.StripPrefix(path, fs)
	http.Handle(path, web.InterceptFatal(web.InterceptCORS(react.InterceptReact(folder, sp))))

}
