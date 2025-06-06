package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	Note         int
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
	http.HandleFunc("/logout", logoutPage)
	http.HandleFunc("/cocktails", cocktailsPage)
	http.HandleFunc("/cocktail", cocktailPage)
	http.HandleFunc("/creation", creationPage)
	http.HandleFunc("/user/profile", userProfileHandler)
	http.HandleFunc("/accueil", accueilHandler)
	http.HandleFunc("/profil", profilPage)

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
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		tmpl, _ := template.ParseFiles("html/login.html")
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")

		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&credentials); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "Format de données invalide",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		isValid, userID, err := db.CheckUser(database, credentials.Email, credentials.Password)
		if err != nil || !isValid {
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]interface{}{
				"success": false,
				"message": "Email ou mot de passe incorrect",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		username, err := db.GetUsernameWithID(database, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]interface{}{
				"success": false,
				"message": "Erreur lors de la récupération du pseudo",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		expiration := time.Now().Add(3 * time.Hour)
		cookie := &http.Cookie{
			Name:     "session",
			Value:    strconv.Itoa(userID),
			Path:     "/",
			Expires:  expiration,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)

		response := map[string]interface{}{
			"success": true,
			"message": "Connexion réussie",
			"pseudo":  username,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	response := map[string]interface{}{
		"success": false,
		"message": "Méthode non autorisée",
	}
	json.NewEncoder(w).Encode(response)
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		tmpl, _ := template.ParseFiles("html/register.html")
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")

		var credentials struct {
			Pseudo   string `json:"pseudo"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&credentials); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "Format de données invalide",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		exists, err := db.UserExists(database, credentials.Pseudo, credentials.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]interface{}{
				"success": false,
				"message": "Erreur lors de la vérification de l'utilisateur",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		if exists {
			w.WriteHeader(http.StatusConflict)
			response := map[string]interface{}{
				"success": false,
				"message": "Le pseudo ou l'email existe déjà",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		err = db.CreateUser(database, credentials.Pseudo, credentials.Email, credentials.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]interface{}{
				"success": false,
				"message": "Erreur lors de la création de l'utilisateur",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		userID, err := db.GetUserID(database, credentials.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]interface{}{
				"success": false,
				"message": "Erreur lors de la récupération de l'ID utilisateur",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		expiration := time.Now().Add(3 * time.Hour)
		cookie := &http.Cookie{
			Name:     "session",
			Value:    strconv.Itoa(userID),
			Path:     "/",
			Expires:  expiration,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)

		response := map[string]interface{}{
			"success": true,
			"message": "Compte créé avec succès",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	response := map[string]interface{}{
		"success": false,
		"message": "Méthode non autorisée",
	}
	json.NewEncoder(w).Encode(response)
}

func accountPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	// Vérifier s'il y a des messages d'erreur ou de succès dans l'URL
	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		switch errorMsg {
		case "update_failed":
			data.Error = "La mise à jour des informations a échoué. Veuillez réessayer."
		case "empty_fields":
			data.Error = "Tous les champs obligatoires doivent être remplis."
		case "invalid_email":
			data.Error = "Format d'email invalide. Veuillez entrer une adresse email valide."
		default:
			data.Error = "Une erreur s'est produite."
		}
	}

	if successMsg := r.URL.Query().Get("success"); successMsg != "" {
		switch successMsg {
		case "updated":
			data.Success = "Vos informations ont été mises à jour avec succès."
		default:
			data.Success = "Opération réussie."
		}
	}

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
		data.IsLogged = true
		data.Note = user.Note
		creationDate, err := time.Parse("2006-01-02T15:04:05Z07:00", user.CreationDate)
		if err == nil {
			data.CreationDate = creationDate.Format("02/01/2006")
		} else {
			data.CreationDate = "Date invalide"
		}
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

			// Validation des champs
			if newUsername == "" || newEmail == "" {
				http.Redirect(w, r, "/account?error=empty_fields", http.StatusSeeOther)
				return
			}

			// Valider le format de l'email
			if !strings.Contains(newEmail, "@") || !strings.Contains(newEmail, ".") {
				http.Redirect(w, r, "/account?error=invalid_email", http.StatusSeeOther)
				return
			}

			// La fonction UpdateUserInfo gère déjà le cas où le mot de passe est vide
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

func logoutPage(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func creationPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, _ := strconv.Atoi(cookie.Value)

	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("html/creation.html")
		if err != nil {
			http.Error(w, "Erreur lors du chargement de la page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
	} else if r.Method == http.MethodPost {
		name := r.FormValue("name")
		ingredients := r.FormValue("ingredients")
		recette := r.FormValue("recipe")
		ustensile := r.FormValue("utensils")
		tempsPreparation, _ := strconv.Atoi(r.FormValue("preparation_time"))

		if name == "" || ingredients == "" || recette == "" || ustensile == "" || tempsPreparation <= 0 {
			data.Error = "Tous les champs sont obligatoires."
			tmpl, _ := template.ParseFiles("html/creation.html")
			tmpl.Execute(w, data)
			return
		}

		err := db.CreateCocktail(database, userID, name, ingredients, recette, ustensile, tempsPreparation)
		if err != nil {
			data.Error = "Erreur lors de la création du cocktail."
			tmpl, _ := template.ParseFiles("html/creation.html")
			tmpl.Execute(w, data)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func cocktailsPage(w http.ResponseWriter, r *http.Request) {
	cocktails, err := db.GetCocktails(database)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors de la récupération des cocktails : %v", err), http.StatusInternalServerError)
		return
	}

	// Structure temporaire pour le JSON
	type CocktailJSON struct {
		ID               int      `json:"id"`
		IDCreator        int      `json:"idCreator"`
		Name             string   `json:"name"`
		Ingredients      []string `json:"ingredients"`
		Recette          string   `json:"recette"`
		Ustensile        []string `json:"ustensile"`
		TempsPreparation int      `json:"tempsPreparation"`
		CreatorUsername  string   `json:"creatorUsername"`
	}

	// Convertir les cocktails pour le format JSON
	cocktailsJSON := make([]CocktailJSON, 0, len(cocktails))

	for _, c := range cocktails {
		username, _ := db.GetUsernameWithID(database, c.IDCreator)

		cocktailJSON := CocktailJSON{
			ID:               c.ID,
			IDCreator:        c.IDCreator,
			Name:             c.Name,
			Ingredients:      strings.Split(c.Ingredients, ","),
			Recette:          c.Recette,
			Ustensile:        strings.Split(c.Ustensile, ","),
			TempsPreparation: c.TempsPreparation,
			CreatorUsername:  username,
		}

		cocktailsJSON = append(cocktailsJSON, cocktailJSON)
	}

	// Vérifier si le format JSON est demandé
	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"success":   true,
			"cocktails": cocktailsJSON,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Sinon, afficher la page HTML
	tmpl := template.Must(template.ParseFiles("html/cocktails.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Cocktails": cocktailsJSON,
	})
}

func cocktailPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	type Cocktail struct {
		ID               int
		IDCreator        int
		Name             string
		Ingredients      []string
		Recette          string
		Ustensile        []string
		TempsPreparation int
		CreatorUsername  string // Ajouter un champ pour le username du créateur
	}

	query := `
        SELECT id, idcreator, name, ingredients, recette, ustensile, temps_preparation
        FROM cocktail
        WHERE id = ? 
    `
	var cocktail Cocktail
	var ingredientsStr, ustensileStr string

	// Récupérer les informations du cocktail
	err = database.QueryRow(query, id).Scan(
		&cocktail.ID,
		&cocktail.IDCreator,
		&cocktail.Name,
		&ingredientsStr,
		&cocktail.Recette,
		&ustensileStr,
		&cocktail.TempsPreparation,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Cocktail non trouvé", http.StatusNotFound)
		log.Println("Cocktail non trouvé avec ID:", id)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors de la récupération du cocktail : %v", err), http.StatusInternalServerError)
		log.Println("Erreur lors de la récupération du cocktail:", err)
		return
	}

	creatorUsername, _ := db.GetUsernameWithID(database, cocktail.IDCreator)

	cocktail.CreatorUsername = creatorUsername

	cocktail.Ingredients = strings.Split(ingredientsStr, ",")
	cocktail.Ustensile = strings.Split(ustensileStr, ",")

	tmpl := template.Must(template.ParseFiles("html/cocktail.html"))
	err = tmpl.Execute(w, cocktail)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur lors du rendu du template : %v", err), http.StatusInternalServerError)
		return
	}
}

func profilPage(w http.ResponseWriter, r *http.Request) {
	// Vérifier si l'utilisateur est connecté
	_, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Afficher la page de profil
	tmpl, _ := template.ParseFiles("html/profil.html")
	tmpl.Execute(w, nil)
}

func userProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si l'utilisateur est connecté
	cookie, err := r.Cookie("session")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"success": false,
			"message": "Non authentifié",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Récupérer l'ID de l'utilisateur
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"success": false,
			"message": "Session invalide",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Récupérer les informations de l'utilisateur
	user, err := db.GetUserInfo(database, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"success": false,
			"message": "Erreur lors de la récupération des informations de l'utilisateur",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Renvoyer les informations
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"pseudo":  user.Username,
		"email":   user.Email,
	}
	json.NewEncoder(w).Encode(response)
}

func accueilHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si l'utilisateur est connecté
	_, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Afficher la page d'accueil
	w.Header().Set("Content-Type", "text/html")
	tmpl, _ := template.ParseFiles("html/accueil.html")
	tmpl.Execute(w, nil)
}
