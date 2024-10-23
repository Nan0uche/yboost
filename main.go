// main.go
package main

import (
	"fmt"
	"log"

	"example.com/project/db"
)

func main() {
	fmt.Println("Initialisation de la base de données...")

	// Initialiser la base de données
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données : %v", err)
	}
	defer db.CloseDB(database)

	// Ici, vous pouvez ajouter du code pour insérer, mettre à jour ou lire des données
}
