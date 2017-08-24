package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"../confreader"
	"../contexts/userlogin"
	"../models/user"
	"../rpnr"
	"../workers"
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

var conf serverConfig

var wp IPool

const requestWaitInQueueTimeout = time.Millisecond * 100

var ctx = context.Background()

// ab -c10 -n20 localhost:8000/hello

func checkLogin(w http.ResponseWriter, r *http.Request) {
	val, ok := ctx.Value(userlogin.UserLoginKey).(string)
	fmt.Println(val, ok)
	if !ok {
		http.Redirect(w, r, "/login", 301)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		tmpl := template.Must(template.ParseFiles("assets/login.html"))
		fmt.Println("method:", r.Method)
		if r.Method == http.MethodGet {
			tmpl.Execute(w, nil)
		} else {
			token, _ := userlogin.LoginUser(r)
			ctx = userlogin.NewContext(ctx, token)
			http.Redirect(w, r, "/", 301)
		}
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		tmpl := template.Must(template.ParseFiles("assets/registration.html"))
		fmt.Println("method:", r.Method)
		if r.Method == http.MethodGet {
			tmpl.Execute(w, nil)
		} else {
			if r.FormValue("password") != r.FormValue("confirm") {
				tmpl.Execute(w, struct{ Res string }{Res: "Passwords don't match"})
				return nil
			}
			u := user.User{Username: r.FormValue("username"), Password: r.FormValue("password")}
			err := u.Create()
			if err != nil {
				tmpl.Execute(w, struct{ Res string }{Res: "Can't create user"})
				return nil
			}
			token, _ := userlogin.LoginUser(r)
			ctx = userlogin.NewContext(ctx, token)
			http.Redirect(w, r, "/", 301)
		}
		return nil
	}, requestWaitInQueueTimeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %s!\n", err), 500)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		tmpl := template.Must(template.ParseFiles("assets/upload.html"))
		fmt.Println("method:", r.Method)
		fmt.Println("in root")
		checkLogin(w, r)
		if r.Method == http.MethodGet {

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
	if err := confreader.ReadConfig("server", &conf); err != nil {
		fmt.Println(err)
	}
	fmt.Println(conf)
	wp = workers.NewPool(conf.Workers)
	wp.Run()
}

//RunHTTPServer - runs http server on address addr
func RunHTTPServer() error {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", regHandler)
	return http.ListenAndServe(conf.Server, nil)
}
