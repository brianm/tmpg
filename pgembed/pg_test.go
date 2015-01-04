package pgembed

import (
	"database/sql"
	_ "gopkg.in/jackc/pgx.v2/stdlib"
	"os"

	"fmt"
	"testing"
)

func TestStart(t *testing.T) {
	pg := &PgEmbed{}
	err := pg.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Stop()

	if pg.DataDir == "" {
		t.Errorf("DataDir should be set after starting")
	}

	if pg.Port == 0 {
		t.Errorf("Port should be set after starting")
	}

	if pg.Superuser != "postgres" {
		t.Errorf("Superuser should be 'postgres' after start")
	}
}

func TestNetworkConnect(t *testing.T) {
	pg := &PgEmbed{}
	err := pg.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Stop()

	conn := fmt.Sprintf("postgres://postgres@localhost:%d/postgres", pg.Port)
	db, err := sql.Open("pgx", conn)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`create table things (
                        id int primary key,
                        name text )`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`insert into things (id, name) values 
                        (1, 'Brian'), (2, 'Matt')`)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query("select name from things order by id")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var name string
	if !rows.Next() {
		t.Fatalf("expected first row!")
	}
	if err = rows.Scan(&name); err != nil {
		t.Errorf("error scanning for Brian: %s", err)
	}
	if name != "Brian" {
		t.Fatalf("expected Brian got %s", name)
	}

	if !rows.Next() {
		t.Fatalf("expected second row!")
	}
	if err = rows.Scan(&name); err != nil {
		t.Fatalf("error scanning for Matt: %s", err)
	}
	if name != "Matt" {
		t.Fatalf("expected Matt, got %s", name)
	}

	if rows.Next() {
		t.Fatalf("expected to be out of rows, was not!")
	}

}
