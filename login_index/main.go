<!DOCTYPE html>
<html>
<head>
        <title>Login Successful</title>
        <style>
                body {
                        font-family: Arial, sans-serif;
                        text-align: center;
                }

                .success-message {
                        color: #3498DB; /* updated color code */
                        font-size: 24px;
                        margin-top: 50px;
                }
        </style>
</head>
<body>
        <div class="success-message">
                <h1>Login Successful!</h1>
                <p>Welcome to our application. You have successfully logged in.</p>
        </div>
</body>
</html>
root@ip-172-31-60-105:/home/ubuntu/blue# cat main.go
package main

import (
        "fmt"
        "net/http"
)

func main() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                http.ServeFile(w, r, "index.html")
        })

        fmt.Println("Server started on port 8080")
        http.ListenAndServe(":8080", nil)
}
