package main

/*
import "path/filepath"

import (
	"database/sql"
	"log"
)

func workDb(appPath string) {
	// Формируем путь к файлу базы данных
	dbFile := filepath.Join(filepath.Dir(appPath), "./scheduler.db")

	// Открываем базу данных
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем таблицу и индекс, если они не существуют
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT,
            title TEXT,
            comment TEXT,
            repeat TEXT
        );
        CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
    `)

	if err != nil {

		log.Fatal(err)
	}
}
*/
