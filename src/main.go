package main

import (
	"net/http"
	"log"
	"sync"
	"html/template"
	"path/filepath"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

type templateHandler struct {
	once sync.Once
	filename string
	templ *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){

	// 1 раз инициализируем template
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil{
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	// дохуя раз выполняем template
	t.templ.Execute(w, data)
}

/*Gomniauth requires the SetSecurityKey call because it sends state data between the client and server along with a signature checksum,
which ensures that the state values haven't been tempered with while transmitting.
The security key is used when creating the hash in a way that it is almost impossible to recreate the same hash without knowing the exact security key.*/
func main(){
	gomniauth.SetSecurityKey("123456789")
	gomniauth.WithProviders(
		github.New("9536371b7bd0e5dba76e", "8b1cf566843074c98c89b911b3ab55bdc9a44d94","http://localhost:8080/auth/callback/github"),
		google.New("239445132641-ls2a8hbmedjeu0s5i1ceb9se5beglnfg.apps.googleusercontent.com", "OH5y-7Qtrc9FdbO1obyZix3G", "http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	http.Handle("/chat", MustAuth(&templateHandler{filename:"chat.html"}))
	http.Handle("/login", &templateHandler{filename:"login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	http.HandleFunc("/logout", func(w http.ResponseWriter, req *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:"auth",
			Value:"",
			Path:"/",
			MaxAge:-1,
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	go r.run()
	if err := http.ListenAndServe(":8080", nil); err != nil{
		log.Fatal("ListenAndServe: ", err)
	}
}
