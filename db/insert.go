package db

import (
	"database/sql"
	"fmt"
)

func AddCocktail(db *sql.DB, idcreator int, name, ingredients, recette, ustensile string, tempsPreparation int) error {
	insertCocktailSQL := `
    INSERT INTO cocktail (idcreator, name, ingredients, recette, ustensile, temps_preparation)
    VALUES (?, ?, ?, ?, ?, ?);`

	_, err := db.Exec(insertCocktailSQL, idcreator, name, ingredients, recette, ustensile, tempsPreparation)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout du cocktail : %w", err)
	}

	fmt.Println("Cocktail ajouté avec succès.")
	return nil
}

func AddAccount(db *sql.DB, username, email, password, creationDate string) error {
	insertAccountSQL := `
    INSERT INTO account (username, email, password, creation_date)
    VALUES (?, ?, ?, ?);`

	_, err := db.Exec(insertAccountSQL, username, email, password, creationDate)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout du compte : %w", err)
	}

	fmt.Println("Compte ajouté avec succès.")
	return nil
}

func AddAvis(db *sql.DB, idCocktail, idUser, note int, commentaire, dateAvis string) error {
	insertAvisSQL := `
    INSERT INTO avis (id_cocktail, id_user, note, commentaire, date_avis)
    VALUES (?, ?, ?, ?, ?);`

	_, err := db.Exec(insertAvisSQL, idCocktail, idUser, note, commentaire, dateAvis)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ajout de l'avis : %w", err)
	}

	fmt.Println("Avis ajouté avec succès.")
	return nil
}
