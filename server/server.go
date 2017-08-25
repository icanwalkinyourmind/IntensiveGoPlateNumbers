package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"../confreader"
	"../contexts/loggedin"
	"../rpnr"
	"../workers"
	"./controller"
)

type serverConfig struct {
	Server  string
	Workers int
}

//IPool - pool of routines
type IPool interface {
	Size() int
	Run()
	AddTaskSyncTimed(f workers.Func, timeout time.Duration) (interface{}, error)
}

const (
	requestWaitInQueueTimeout = time.Millisecond * 100
	solt                      = "i_love_perl"
)

//html templates
var index = template.Must(template.ParseFiles("assets/index.html"))
var login = template.Must(template.ParseFiles("assets/login.html"))
var registration = template.Must(template.ParseFiles("assets/registration.html"))

//very important stuff
var sessions = make(map[string]string)
var conf serverConfig
var wp IPool

func loginHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		fmt.Println("method:", r.Method)
		if r.Method == http.MethodGet {
			login.Execute(w, nil)
		} else {
			err := controller.LoginUser(w, r, &sessions)
			if err != nil {
				login.Execute(w, nil)
			} else {
				index.Execute(w, nil)
			}
		}
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		tmpl := registration
		fmt.Println("method:", r.Method)
		if r.Method == http.MethodGet {
			tmpl.Execute(w, nil)
		} else {
			err := controller.RegisterUser(w, r)
			if err != nil {
				tmpl.Execute(w, struct{ Res string }{Res: err.Error()})
			} else {
				http.Redirect(w, r, "/login", 301)
			}
		}
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		fmt.Println("method:", r.Method)
		fmt.Println("in root")
		u, ok := loggedin.FromContext(r.Context())
		if !ok {
			http.Redirect(w, r, "/login", 301)
		}
		if r.Method == http.MethodGet {

			index.Execute(w, struct{ Res, Username string }{Res: "", Username: u.Username})

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
			index.Execute(w, struct{ Res string }{Res: "Result: " + result + er})
			//
		}
		return nil
	}, requestWaitInQueueTimeout)

	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		u, ok := loggedin.FromContext(r.Context())
		http.Error(w, fmt.Sprintf("error: %s!\n", ok), 500)
		fmt.Println("in logout " + u.Username)
		if !ok {
			http.Redirect(w, r, "/login", 301)
			return nil
		}
		fmt.Println("after redirect")
		delete(sessions, u.Username)
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func init() {
	if err := confreader.ReadConfig("server", &conf); err != nil {
		fmt.Println(err)
	}
	fmt.Println(conf)
	wp = workers.NewPool(conf.Workers)
	wp.Run()
}

//RunHTTPServer - runs http server on address addr
func RunHTTPServer() error {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/register", regHandler)
	return http.ListenAndServe(conf.Server, loggedin.AddLoginContext(mux, &sessions))
}
