package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func main() {
	entry()
}

func entry() {
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.New("")
		t = t.Funcs(template.FuncMap{
			"count": func(n int) []int {
				var a []int
				for i := 0; i < n; i++ {
					a = append(a, i)
				}
				return a
			},
		})
		t, err := t.ParseGlob("web/*.tmpl")
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := t.ExecuteTemplate(w, "index.tmpl", nil); err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("/install", func(w http.ResponseWriter, r *http.Request) {
		t := template.New("")
		t, err := t.ParseGlob("web/*.tmpl")
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := t.ExecuteTemplate(w, "install.tmpl", nil); err != nil {
			fmt.Println(err)
		}

	})

	fmt.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println(err)
	}
}
