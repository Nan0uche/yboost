package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const dbPath = "./db/cocktail.db"

type User struct {
	ID           int
	Username     string
	Email        string
	CreationDate string
	Note         int
}

type Cocktail struct {
	ID               int
	IDCreator        int
	Name             string
	Ingredients      string
	Recette          string
	Ustensile        string
	TempsPreparation int
}

func InitDB() (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la création du répertoire : %w", err)
		}
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base de données : %w", err)
	}

	createTableCocktail := `
    CREATE TABLE IF NOT EXISTS cocktail (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        idcreator INTEGER NOT NULL,
        name TEXT NOT NULL,
        ingredients TEXT NOT NULL,
        recette TEXT NOT NULL,
        ustensile TEXT NOT NULL,
        temps_preparation INTEGER NOT NULL
    );`

	_, err = db.Exec(createTableCocktail)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la table Cocktail : %w", err)
	}

	createTableAccount := `
    CREATE TABLE IF NOT EXISTS account (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        creation_date TEXT NOT NULL,
		note INTEGER NOT NULL DEFAULT 0
    );`

	_, err = db.Exec(createTableAccount)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la table account : %w", err)
	}

	createTableAvis := `
    CREATE TABLE IF NOT EXISTS avis (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        id_cocktail INTEGER NOT NULL,
        id_user INTEGER NOT NULL,
        note INTEGER NOT NULL,
        commentaire TEXT NOT NULL,
        date_avis TEXT NOT NULL,
        FOREIGN KEY (id_cocktail) REFERENCES cocktail(id) ON DELETE CASCADE,
        FOREIGN KEY (id_user) REFERENCES account(id) ON DELETE CASCADE
    );`

	_, err = db.Exec(createTableAvis)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la table avis : %w", err)
	}

	return db, nil
}

func CloseDB(db *sql.DB) {
	db.Close()
}

func GetUserInfo(db *sql.DB, id int) (User, error) {
	var user User
	query := "SELECT id, username, email, creation_date, note FROM account WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreationDate, &user.Note)
	if err != nil {
		return user, err
	}
	return user, nil
}

func UpdateUserInfo(db *sql.DB, userID int, newUsername, newEmail, newPassword string) error {
	var query string
	var args []interface{}

	if newPassword == "" {
		query = "UPDATE account SET username = ?, email = ? WHERE id = ?"
		args = []interface{}{newUsername, newEmail, userID}
	} else {
		query = "UPDATE account SET username = ?, email = ?, password = ? WHERE id = ?"
		args = []interface{}{newUsername, newEmail, newPassword, userID}
	}

	_, err := db.Exec(query, args...)
	return err
}

func CheckUser(db *sql.DB, email, password string) (bool, int, error) {
	var id int
	var hashedPassword string

	query := "SELECT id, password FROM account WHERE email = ?"
	err := db.QueryRow(query, email).Scan(&id, &hashedPassword)
	if err != nil {
		return false, 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return false, 0, nil
	}

	return true, id, nil
}

func UserExists(db *sql.DB, username, email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM account WHERE username = ? OR email = ?"
	err := db.QueryRow(query, username, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func CreateUser(db *sql.DB, username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	creationDate := time.Now().Format(time.RFC3339)
	query := "INSERT INTO account (username, email, password, creation_date) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(query, username, email, hashedPassword, creationDate)
	return err
}

func GetUserID(db *sql.DB, email string) (int, error) {
	var id int
	query := `SELECT id FROM account WHERE email = ?`
	err := db.QueryRow(query, email).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetUsernameWithID(db *sql.DB, id int) (string, error) {
	var username string
	query := `SELECT username FROM account WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&username)
	if err != nil {
		return "Inconnu", err
	}
	return username, nil
}

func GetCocktails(db *sql.DB) ([]Cocktail, error) {
	query := `
		SELECT id, idcreator, name, ingredients, recette, ustensile, temps_preparation
		FROM cocktail
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des cocktails : %w", err)
	}
	defer rows.Close()

	var cocktails []Cocktail

	for rows.Next() {
		var c Cocktail
		err := rows.Scan(&c.ID, &c.IDCreator, &c.Name, &c.Ingredients, &c.Recette, &c.Ustensile, &c.TempsPreparation)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la lecture des données : %w", err)
		}
		cocktails = append(cocktails, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur finale lors de la récupération des données : %w", err)
	}

	return cocktails, nil
}

func CreateCocktail(db *sql.DB, idcreator int, name, ingredients, recette, ustensile string, tempsPreparation int) error {
	query := `INSERT INTO cocktail (idcreator, name, ingredients, recette, ustensile, temps_preparation) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, idcreator, name, ingredients, recette, ustensile, tempsPreparation)
	return err
}

func CreateAvis(db *sql.DB, cocktailid int, creatorcocktailid, userid, note, commentaire string) error {
	creationDate := time.Now().Format(time.RFC3339)
	query := `INSERT INTO avis (cocktailid, creatorcocktailid, userid, note, commentaire, creationDate) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, cocktailid, creatorcocktailid, userid, note, commentaire, creationDate)
	return err
}
