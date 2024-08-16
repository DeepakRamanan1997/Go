package main

import (
        "log"
        "net/http"
)

func main() {
        http.HandleFunc("/", loginHandler)

        log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" {
                http.ServeFile(w, r, "login.html")
                return
        }

        username := r.FormValue("username")
        password := r.FormValue("password")

        if username == "admin" && password == "admin" {
                http.SetCookie(w, &http.Cookie{Name: "logged_in", Value: "true"})
                http.Redirect(w, r, "http://internal-private-1132681189.ap-south-1.elb.amazonaws.com", http.StatusFound)
                return
        }

        http.ServeFile(w, r, "login.html")
}
