package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	db "github.com/anshul393/snippetbox/db/sqlc"
	"github.com/go-playground/form/v4"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	dtb            db.Querier
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	debug          bool
}

func main() {

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	addr := flag.String("addr", ":4000", "HTTP network address")

	host := flag.String("db_host", "localhost", "hostname for database connection")
	port := flag.Int("db_port", 5432, "port-number for database connection")
	user := flag.String("db_user", "localuser", "user for database connection")
	password := flag.String("db_password", "password", "password for database connection")
	dbname := flag.String("db_name", "snippetbox", "database to connect")
	sslmode := flag.String("ssl_mode", "disable", "sslmode for connection") // or "require" for SSL/TLS
	debug := flag.Bool("debug", false, "to enable the debug mode")

	flag.Parse()

	// Create the DSN string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		*host, *port, *user, *password, *dbname, *sslmode)

	dbconn, err := openDB(dsn)
	if err != nil {
		errLog.Fatal(err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	sessionManager := scs.New()
	sessionManager.Store = redisstore.New(pool)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errLog,
		infoLog:        infoLog,
		dtb:            db.New(dbconn),
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		debug:          *debug,
	}

	defer dbconn.Close()

	tlsConfig := tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errLog,
		Handler:      app.routes(),
		TLSConfig:    &tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Print("Starting server on ", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errLog.Fatal(err)

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
