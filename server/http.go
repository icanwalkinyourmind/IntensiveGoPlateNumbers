package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/icanwalkinyourmind/IntensiveGoPlateNumbers/confreader"
	"github.com/icanwalkinyourmind/IntensiveGoPlateNumbers/rpnr"
	"github.com/icanwalkinyourmind/IntensiveGoPlateNumbers/workers"
)

const configFile = "../config.yaml"

type serverConfig struct {
	Server     string
	nOfWorkers int `yaml:"n_of_workers"`
}

//IPool - pool of routines
type IPool interface {
	Size() int
	Run()
	AddTaskSyncTimed(f workers.Func, timeout time.Duration) (interface{}, error)
}

var conf serverConfig

var wp IPool

const requestWaitInQueueTimeout = time.Millisecond * 100

// ab -c10 -n20 localhost:8000/hello

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		tmpl := template.Must(template.ParseFiles("assets/upload.html"))
		fmt.Println("method:", r.Method)
		if r.Method == "GET" {

			tmpl.Execute(w, struct{ Res string }{Res: "Result: "})

		} else {
			r.ParseMultipartForm(32 << 20)
			file, handler, err := r.FormFile("uploadfile")
			if err != nil {
				fmt.Println(err)
				return nil
			}
			defer file.Close()

			f, err := os.OpenFile("./"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			defer f.Close()
			io.Copy(f, file)
			//
			result, er := rpnr.GetPlateNumber(handler.Filename)
			tmpl.Execute(w, struct{ Res string }{Res: "Result: " + result + er})
			//
		}
		return nil
	}, requestWaitInQueueTimeout)

	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func init() {
	confreader.ReadConfig(configFile, conf)
	wp = workers.NewPool(conf.nOfWorkers)
	wp.Run()
}

//RunHTTPServer - runs http server on address addr
func RunHTTPServer() error {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.HandleFunc("/", rootHandler)
	return http.ListenAndServe(conf.Server, nil)
}
