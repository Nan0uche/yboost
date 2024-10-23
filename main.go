package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
	"yboost/db"
)

type PageData struct {
	Error        string
	Success      string
	IsLogged     bool
	Username     string
	Email        string
	CreationDate string
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
	http.HandleFunc("/account", accountPage)
	http.HandleFunc("/update", updatePage)

	log.Println("Le serveur est en cours d'exécution sur le port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur : %v", err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	var username string

	if err == nil {
		userID, parseErr := strconv.Atoi(cookie.Value)
		if parseErr != nil {
			log.Printf("Erreur lors de la conversion de l'ID utilisateur depuis le cookie : %v", parseErr)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		user, userErr := db.GetUserInfo(database, userID)
		if userErr != nil {
			log.Printf("Erreur lors de la récupération des informations de l'utilisateur : %v", userErr)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		username = user.Username
	}

	data := PageData{
		Success:  "Vous êtes connecté en tant que " + username,
		IsLogged: username != "",
	}

	tmpl, err := template.ParseFiles("html/home.html")
	if err != nil {
		http.Error(w, "Erreur lors du chargement de la page", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func loginPage(w http.ResponseWriter, r *http.Request) {

	if _, err := r.Cookie("session"); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := PageData{}

	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("html/login.html")
		tmpl.Execute(w, data)
	} else if r.Method == http.MethodPost {
		email := r.FormValue("username")
		password := r.FormValue("password")
		isValid, userID, err := db.CheckUser(database, email, password)
		if err != nil {
			data.Error = "Erreur lors de la vérification de l'utilisateur."
			tmpl, _ := template.ParseFiles("html/login.html")
			tmpl.Execute(w, data)
			return
		}
		if isValid {
			expiration := time.Now().Add(3 * time.Hour)
			cookie := &http.Cookie{
				Name:    "session",
				Value:   strconv.Itoa(userID),
				Path:    "/",
				Expires: expiration,
			}
			http.SetCookie(w, cookie)
			data.Success = "Bienvenue " + email + "!"
			http.Redirect(w, r, "/account", http.StatusSeeOther)
		} else {
			data.Error = "Identifiants invalides."
		}
		tmpl, _ := template.ParseFiles("html/login.html")
		tmpl.Execute(w, data)
	}
}

func registerPage(w http.ResponseWriter, r *http.Request) {

	if _, err := r.Cookie("session"); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

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
				userID, err := db.GetUserID(database, email)
				if err != nil {
					data.Error = "Erreur lors de la récupération de l'ID utilisateur."
				} else {
					expiration := time.Now().Add(3 * time.Hour)
					cookie := &http.Cookie{
						Name:    "session",
						Value:   strconv.Itoa(userID),
						Path:    "/",
						Expires: expiration,
					}
					http.SetCookie(w, cookie)
					data.Success = "Compte créé avec succès pour " + username + "!"
					http.Redirect(w, r, "/account", http.StatusSeeOther)
					return
				}
			}
		}
		tmpl, _ := template.ParseFiles("html/register.html")
		tmpl.Execute(w, data)
	}
}

func accountPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{}
	cookie, err := r.Cookie("session")
	if err == nil {
		userID, _ := strconv.Atoi(cookie.Value)

		user, err := db.GetUserInfo(database, userID)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		data.Username = user.Username
		data.Email = user.Email
		data.CreationDate = user.CreationDate
		data.IsLogged = true
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, _ := template.ParseFiles("html/account.html")
	tmpl.Execute(w, data)
}

func updatePage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		userID, _ := strconv.Atoi(cookie.Value)

		if r.Method == http.MethodPost {
			newUsername := r.FormValue("username")
			newEmail := r.FormValue("email")
			newPassword := r.FormValue("password")

			err = db.UpdateUserInfo(database, userID, newUsername, newEmail, newPassword)
			if err != nil {
				http.Redirect(w, r, "/account?error=update_failed", http.StatusSeeOther)
				return
			}

			http.Redirect(w, r, "/account?success=updated", http.StatusSeeOther)
			return
		}
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}
