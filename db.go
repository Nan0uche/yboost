package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./db/test.db"

// InitDB initialise la base de données et crée des tables
func InitDB() (*sql.DB, error) {
	// Ouvre une connexion à la base de données SQLite
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base de données : %w", err)
	}

	// Crée une table de test (exemple)
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création des tables : %w", err)
	}

	fmt.Println("Table users créée avec succès ou déjà existante.")
	return db, nil
}

func CloseDB(db *sql.DB) {
	db.Close()
}
