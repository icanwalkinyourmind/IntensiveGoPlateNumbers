package server

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"../rpnr"
	"../workers"
)

type IPool interface {
	Size() int
	Run()
	AddTaskSyncTimed(f workers.Func, timeout time.Duration) (interface{}, error)
}

var wp IPool

const requestWaitInQueueTimeout = time.Millisecond * 100

// ab -c10 -n20 localhost:8000/hello

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := wp.AddTaskSyncTimed(func() interface{} {
		tmpl := template.Must(template.ParseFiles("assets/upload.html"))
		fmt.Println("method:", r.Method)
		if r.Method == "GET" {
			crutime := time.Now().Unix()
			h := md5.New()
			io.WriteString(h, strconv.FormatInt(crutime, 10))

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

func RunHTTPServer(addr string, n_of_workers int) error {
	wp = workers.NewPool(n_of_workers)
	wp.Run()
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.HandleFunc("/", rootHandler)
	return http.ListenAndServe(addr, nil)
}
