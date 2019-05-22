package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pelletier/go-toml"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func loadSettings() (*toml.Tree, error) {
	file, err := os.Open("database.toml")
	if err != nil {
		return nil, err
	}
	fmt.Println("We loaded the file", file.Name())

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, fileInfo.Size())
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}
	config, _ := toml.Load(string(data))
	return config, nil
}

func connectToDatabase() (*sql.DB, error) {
	config, err := loadSettings()
	if err != nil {
		return nil, err
	}

	// Prepare database
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=disable",
		config.Get("database.host"), config.Get("database.port"), config.Get("database.user"), config.Get("database.dbname"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}

func createMigrationsTableIfNew(db *sql.DB) error {
	createVersionTableSQL := `
		CREATE TABLE IF NOT EXISTS migrations (last_migrated INT);
		INSERT INTO migrations (last_migrated) VALUES (0);`
	_, err := db.Exec(createVersionTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func getLastMigratedVersion(db *sql.DB) int {
	const getMigratedSQL = "SELECT last_migrated FROM migrations"

	row := db.QueryRow(getMigratedSQL)
	migrationVersion := struct{ version int }{version: 0}
	err := row.Scan(&migrationVersion.version)
	if err != nil {
		migrationVersion.version = -1
	}

	return migrationVersion.version
}

func runMigrations(db *sql.DB) {
	version := getLastMigratedVersion(db)
	fmt.Println("Running migrations")
	file, err := os.Open("migrations")
	if err != nil {
		log.Fatal(err)
	}

	files, err := file.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}

	var migrationVersion int

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, fileInfo := range files {
		firstPartOfFilename := strings.Split(fileInfo.Name(), "-")

		migrationVersion, err = strconv.Atoi(firstPartOfFilename[0])
		if err != nil || migrationVersion <= version {
			fmt.Println(fileInfo.Name(), "[skipped]")
			continue
		}

		sqlFile, err := os.Open(file.Name() + "/" + fileInfo.Name())
		if err != nil {
			log.Fatal(err)
		}

		data := make([]byte, fileInfo.Size())

		_, err = sqlFile.Read(data)
		if err != nil {
			log.Fatal(err)
		}

		sql := string(data)

		_, err = db.Exec(sql)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(fileInfo.Name(), "[done]")
	}
	db.Exec("UPDATE migrations SET last_migrated = $1", migrationVersion)
}

func main() {
	db, err := connectToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = createMigrationsTableIfNew(db)
	if err != nil {
		log.Fatal(err)
	}

	runMigrations(db)
}
