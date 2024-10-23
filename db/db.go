package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const dbPath = "./db/cocktail.db"

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base de données : %w", err)
	}

	err = InitCocktailTable(db)
	if err != nil {
		return nil, err
	}

	err = InitAccountTable(db)
	if err != nil {
		return nil, err
	}

	err = InitAvisTable(db)
	if err != nil {
		return nil, err
	}

	fmt.Println("Toutes les tables ont été créées avec succès ou sont déjà existantes.")
	return db, nil
}

func InitCocktailTable(db *sql.DB) error {
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

	_, err := db.Exec(createTableCocktail)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la table Cocktail : %w", err)
	}
	fmt.Println("Table 'cocktail' créée avec succès ou déjà existante.")
	return nil
}

func InitAccountTable(db *sql.DB) error {
	createTableAccount := `
    CREATE TABLE IF NOT EXISTS account (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        creation_date TEXT NOT NULL,
		note INTEGER NOT NULL
    );`

	_, err := db.Exec(createTableAccount)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la table account : %w", err)
	}
	fmt.Println("Table 'account' créée avec succès ou déjà existante.")
	return nil
}

func InitAvisTable(db *sql.DB) error {
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

	_, err := db.Exec(createTableAvis)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la table avis : %w", err)
	}
	fmt.Println("Table 'avis' créée avec succès ou déjà existante.")
	return nil
}

func CloseDB(db *sql.DB) {
	db.Close()
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CreateUser(db *sql.DB, username, email, password string) error {
	creationDate := time.Now().Format("2006-01-02 15:04:05")
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO account (username, email, password, creation_date, note) VALUES (?, ?, ?, ?, 0)`,
		username, email, hashedPassword, creationDate)
	return err
}

func CheckUser(db *sql.DB, email, password string) (bool, error) {
	var hashedPassword string
	err := db.QueryRow(`SELECT password FROM account WHERE email = ?`, email).Scan(&hashedPassword)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

func UserExists(db *sql.DB, username, email string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM account WHERE username = ? OR email = ?`, username, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
