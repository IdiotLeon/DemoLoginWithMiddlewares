package main

import (
	"crypto/subtle"
	"fmt"
	"github.com/satori/go.uuid"
	"net/http"
	"sync"
)

const loginPage = "<html><head><title>Login</title></head></body><form action=\"login\" method=\"post\"> <input type=\"password\" name=\"password\"/><input type=\"submit\" value=\"login\"/></form></body></html>"

func main() {
	sessionStore = make(map[string]Client)

	http.Handle("/hello", helloWorldHandler{})
	http.Handle("/secureHello", authenticate(helloWorldHandler{}))
	http.HandleFunc("/login", handleLogin)

	http.ListenAndServe(":3000", nil)
}

type helloWorldHandler struct {
}

func (h helloWorldHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World!!!")
}

type authenticationMiddleware struct {
	wrappedHandler http.Handler
}

func (h authenticationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		if err != http.ErrNoCookie {
			fmt.Fprint(w, err)
			return
		} else {
			err = nil
		}
	}

	// Unless the cookie exists, if it's saved in our map;
	// if it's not, we will later generate a new one
	var present bool
	var client Client
	if cookie != nil {
		storageMutext.RLock()
		client, present = sessionStore[cookie.Value]
		storageMutext.RUnlock()
	} else {
		present = false
	}

	// if the cookie wasn't present, then we can generate a new one
	if present == false {
		uuid, _ := uuid.NewV4()
		cookie = &http.Cookie{
			Name:  "session",
			Value: uuid.String(),
		}
		client = Client{false}
		storageMutext.Lock()
		sessionStore[cookie.Value] = client
		storageMutext.Unlock()
	}

	// To set the cookie to our response writer:
	// 1. If the client isn't logged in, to send him to the login page;
	// 2. If one is logged in, to send him what he wants
	http.SetCookie(w, cookie)
	if client.loggedIn == false {
		fmt.Fprint(w, loginPage)
		return
	}
	if client.loggedIn == true {
		h.wrappedHandler.ServeHTTP(w, r)
		return
	}
}

func authenticate(h http.Handler) authenticationMiddleware {
	return authenticationMiddleware{h}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// the main part of the middleware
	cookie, err := r.Cookie("session")
	if err != nil {
		if err != http.ErrNoCookie {
			fmt.Fprint(w, err)
			return
		} else {
			err = nil
		}
	}
	var present bool
	var client Client
	if cookie != nil {
		storageMutext.RLock()
		client, present = sessionStore[cookie.Value]
		storageMutext.RUnlock()
	} else {
		present = false
	}

	if present == false {
		uuid, _ := uuid.NewV4()
		cookie = &http.Cookie{
			Name:  "session",
			Value: uuid.String(),
		}
		client = Client{false}
		storageMutext.Lock()
		sessionStore[cookie.Value] = client
		storageMutext.Unlock()
	}
	http.SetCookie(w, cookie)
	err = r.ParseForm()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	if subtle.ConstantTimeCompare([]byte(r.FormValue("password")), []byte("password123")) == 1 {
		// To login user
		client.loggedIn = true
		fmt.Fprintln(w, "Tahnk you for your logging in.")
		storageMutext.Lock()
		sessionStore[cookie.Value] = client
		storageMutext.Unlock()
	} else {
		fmt.Fprintln(w, "Wrong password.")
	}
}

var sessionStore map[string]Client
var storageMutext sync.RWMutex

type Client struct {
	loggedIn bool
}
