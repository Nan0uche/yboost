package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"yboost/db"
)

type PageData struct {
	Error   string
	Success string
}

var database *sql.DB

func main() {
	log.Println("Initialisation de la base de données...")

	var err error
	database, err = db.InitDB()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données : %v", err)
	}
	defer db.CloseDB(database)

	http.HandleFunc("/", homePage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/register", registerPage)

	log.Println("Le serveur est en cours d'exécution sur le port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur : %v", err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/home.html")
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("html/login.html")
		tmpl.Execute(w, data)
	} else if r.Method == http.MethodPost {
		email := r.FormValue("username")
		password := r.FormValue("password")
		isValid, err := db.CheckUser(database, email, password)
		if err != nil {
			data.Error = "Erreur lors de la vérification de l'utilisateur."
			tmpl, _ := template.ParseFiles("html/login.html")
			tmpl.Execute(w, data)
			return
		}
		if isValid {
			data.Success = "Bienvenue " + email + "!"
		} else {
			data.Error = "Identifiants invalides."
		}
		tmpl, _ := template.ParseFiles("html/login.html")
		tmpl.Execute(w, data)
	}
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("html/register.html")
		tmpl.Execute(w, data)
	} else if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		exists, err := db.UserExists(database, username, email)
		if err != nil {
			data.Error = "Erreur lors de la vérification de l'utilisateur."
			tmpl, _ := template.ParseFiles("html/register.html")
			tmpl.Execute(w, data)
			return
		}
		if exists {
			data.Error = "Le nom d'utilisateur ou l'email existe déjà."
		} else {
			err = db.CreateUser(database, username, email, password)
			if err != nil {
				data.Error = "Erreur lors de la création de l'utilisateur."
			} else {
				data.Success = "Compte créé avec succès pour " + username + "!"
			}
		}
		tmpl, _ := template.ParseFiles("html/register.html")
		tmpl.Execute(w, data)
	}
}
