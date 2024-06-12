package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	addr := flag.String("addr", ":3000", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := NewPostgresDB()
	if err != nil {
		errorLog.Fatal(err)
	}

	if err := db.Init(); err != nil {
		errorLog.Fatal(err)
	}

	server := NewApiServer(*addr, db, infoLog, errorLog)
	server.Run()
}
