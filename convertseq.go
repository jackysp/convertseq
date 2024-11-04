package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mode           = flag.String("mode", "", "Mode of operation: sync or restore")
	syncUser       = flag.String("syncUser", "admin", "Sync user")
	syncIP         = flag.String("syncIP", "127.0.0.1", "Sync IP address")
	syncPort       = flag.Int("syncPort", 4000, "Sync port")
	syncPasswd     = flag.String("syncPasswd", "admin", "Sync password")
	syncInterval   = flag.Duration("syncInterval", 5*time.Second, "Sync interval")
	restoreUser    = flag.String("restoreUser", "admin", "Restore user")
	restoreIP      = flag.String("restoreIP", "127.0.0.1", "Restore IP address")
	restorePort    = flag.Int("restorePort", 4000, "Restore port")
	restorePasswd  = flag.String("restorePasswd", "admin", "Restore password")
	restoreWorkers = flag.Int("restoreWorkers", 5, "Number of workers for restore operation")
	//added Schema parameter by Swee
	syncSchema	   = flag.String("syncSchema","test","Sync Schema")
	restoreSchema  = flag.String("restoreSchema","test","Restore Schema")
	logFilePath    = flag.String("logFilePath", "error.log", "Path to error log file")
)

func main() {
	flag.Parse()

	// Added by Swee: Set the log output to the file
	logFile, err := os.OpenFile(*logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile) 

	if *mode != "sync" && *mode != "restore" {
		fmt.Println("Usage: go run main.go -mode=<sync|restore>")
		os.Exit(1)
	}

	var dsn string

	switch *mode {
	case "sync":
		//dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/test", *syncUser, *syncPasswd, *syncIP, *syncPort)
		//added Schema parameter by Swee
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/", *syncUser, *syncPasswd, *syncIP, *syncPort)
	case "restore":
		//dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/test", *restoreUser, *restorePasswd, *restoreIP, *restorePort)
		//added Schema parameter by Swee
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/", *restoreUser, *restorePasswd, *restoreIP, *restorePort)
	default:
		fmt.Printf("Invalid mode: %s\n", *mode)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	switch *mode {
	case "sync":
		syncSeq(db,*syncSchema)
	case "restore":
		restoreSeq(db,*restoreSchema)
	}
}

func syncSeq(db *sql.DB , schema string) {
	//add CREATE DATABASE by Muhaira
	_, err1 := db.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS `+ schema +`;`))
	if err1 != nil {
		log.Fatalf("Failed to create database: %v", err1)
	}


	// Create table if not exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS ` + schema + `.sequence_sync (
        schema_name varchar(64) NOT NULL,
        sequence_name varchar(64) NOT NULL,
        current_value BIGINT UNSIGNED NULL,
        create_sql varchar(300) NULL,
        update_time DATETIME NULL,
        PRIMARY KEY (schema_name, sequence_name)
    );`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	for {
		// Insert or update sequence information
		trx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to execute begin statement: %v", err)
		}
		_, err = trx.Exec(`REPLACE INTO ` + schema + `.sequence_sync (schema_name, sequence_name, create_sql)
		SELECT SEQUENCE_SCHEMA, SEQUENCE_NAME,
		CASE
			WHEN CACHE = 0 AND CYCLE = 0 THEN CONCAT('CREATE SEQUENCE ', SEQUENCE_SCHEMA, '.', SEQUENCE_NAME, ' START WITH ', START, ' MINVALUE ', MIN_VALUE, ' MAXVALUE ', MAX_VALUE, ' INCREMENT BY ', INCREMENT, ' NOCACHE NOCYCLE;')
			WHEN CACHE = 1 AND CYCLE = 0 THEN CONCAT('CREATE SEQUENCE ', SEQUENCE_SCHEMA, '.', SEQUENCE_NAME, ' START WITH ', START, ' MINVALUE ', MIN_VALUE, ' MAXVALUE ', MAX_VALUE, ' INCREMENT BY ', INCREMENT, ' CACHE ', CACHE_VALUE, ' NOCYCLE;')
			WHEN CACHE = 0 AND CYCLE = 1 THEN CONCAT('CREATE SEQUENCE ', SEQUENCE_SCHEMA, '.', SEQUENCE_NAME, ' START WITH ', START, ' MINVALUE ', MIN_VALUE, ' MAXVALUE ', MAX_VALUE, ' INCREMENT BY ', INCREMENT, ' NOCACHE CYCLE;')
			WHEN CACHE = 1 AND CYCLE = 1 THEN CONCAT('CREATE SEQUENCE ', SEQUENCE_SCHEMA, '.', SEQUENCE_NAME, ' START WITH ', START, ' MINVALUE ', MIN_VALUE, ' MAXVALUE ', MAX_VALUE, ' INCREMENT BY ', INCREMENT, ' CACHE ', CACHE_VALUE, ' CYCLE;')
		END AS create_sql
		FROM information_schema.sequences;`)
		if err != nil {
			log.Fatalf("Failed to insert or update sequence information: %v", err)
		}

		// Read data from sequence_sync to show table next_row_id and filter only the type is sequence
		rows, err := db.Query("SELECT schema_name, sequence_name FROM " + schema + ".sequence_sync")
		if err != nil {
			log.Fatalf("Failed to query sequence_sync: %v", err)
		}

		for rows.Next() {
			var nextNotCachedValue int64
			var schemaName, sequenceName string
			if err := rows.Scan(&schemaName, &sequenceName); err != nil {
				log.Fatalf("Failed to scan row: %v", err)
			}

			query := fmt.Sprintf("SHOW TABLE `%s`.`%s` NEXT_ROW_ID", schemaName, sequenceName)
			results, err := db.Query(query)
			if err != nil {
				log.Fatalf("Failed to execute query: %v", err)
			}

			for results.Next() {
				var dbName, tableName, columnName, nextGlobalRowID, idType string
				if err := results.Scan(&dbName, &tableName, &columnName, &nextGlobalRowID, &idType); err != nil {
					log.Fatalf("Failed to scan result: %v", err)
				}
				if idType == "SEQUENCE" {
					nextNotCachedValue, _ = strconv.ParseInt(nextGlobalRowID, 10, 64)
				}
			}
			if err := results.Err(); err != nil {
				log.Fatalf("Error iterating over results: %v", err)
			}
			results.Close()

			// Directly execute the update statement
			updateStatement := fmt.Sprintf("UPDATE " + schema +".sequence_sync SET current_value=%d, update_time=NOW() WHERE schema_name='%s' AND sequence_name='%s';", nextNotCachedValue, schemaName, sequenceName)
			_, err = trx.Exec(updateStatement)
			if err != nil {
				log.Fatalf("Failed to execute update statement: %v", err)
			}
		}
		trx.Commit()
		fmt.Printf("All sequences updated at %s.\n", time.Now().Format("2006-01-02 15:04:05"))

		if err := rows.Err(); err != nil {
			log.Fatalf("Error iterating over rows: %v", err)
		}
		rows.Close()

		time.Sleep(*syncInterval)
	}
}

func restoreSeq(db *sql.DB,schema string) {
	// Generate DROP SEQUENCE statements for existing sequences that need to be dropped
	dropStatements := getSQLStatements(db, "SELECT CONCAT('DROP SEQUENCE ', sequences.SEQUENCE_SCHEMA, '.', sequences.SEQUENCE_NAME, ';') FROM information_schema.sequences JOIN " + schema +".sequence_sync ON sequences.SEQUENCE_SCHEMA = sequence_sync.schema_name AND sequences.SEQUENCE_NAME = sequence_sync.sequence_name;")
	if len(dropStatements) == 0 {
		fmt.Println("No sequences to drop.")
	} else {
		fmt.Println("Dropping old sequences...")
		executeSQLStatements(db, dropStatements)
	}

	// Execute restore operations from sequence_sync
	sqlStatements := getSQLStatements(db, "SELECT create_sql FROM " + schema +".sequence_sync;")
	if len(sqlStatements) == 0 {
		fmt.Println("No sequences to restore.")
	} else {
		fmt.Println("Restoring sequences...")
		executeSQLStatements(db, sqlStatements)
	}

	// Setting current value for sequence
	setvalStatements := getSQLStatements(db, "SELECT CONCAT('SELECT setval(', schema_name, '.', sequence_name, ',', current_value, ');') FROM " + schema +".sequence_sync WHERE current_value IS NOT NULL;")
	if len(setvalStatements) == 0 {
		fmt.Println("No sequences to set.")
	} else {
		fmt.Println("Setting sequences...")
		executeSQLStatements(db, setvalStatements)
	}
}

func getSQLStatements(db *sql.DB, query string) []string {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to execute query: %s, error: %v", query, err)
	}
	defer rows.Close()

	var statements []string
	for rows.Next() {
		var statement string
		if err := rows.Scan(&statement); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		statements = append(statements, statement)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating over rows: %v", err)
	}

	return statements
}

func executeSQLStatements(db *sql.DB, statements []string) {
	numWorkers := *restoreWorkers
	jobs := make(chan string, len(statements))
	results := make(chan error, len(statements))

	// Worker function
	worker := func(jobs <-chan string, results chan<- error) {
		for sql := range jobs {
			_, err := db.Exec(sql)
			results <- err
		}
	}

	// Start workers
	for w := 0; w < numWorkers; w++ {
		go worker(jobs, results)
	}

	// Send jobs to workers
	for _, sql := range statements {
		jobs <- sql
	}
	close(jobs)

	// Collect results
	for i := 0; i < len(statements); i++ {
		if err := <-results; err != nil {
			log.Fatalf("Failed to execute SQL statement: %v", err)
		}
	}
}
