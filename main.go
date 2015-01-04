package main

import (
	"github.com/brianm/tmpg/pgembed"
	"gopkg.in/alecthomas/kingpin.v1"

	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
)

var (
	verbose *bool
	uname   *bool
)

func init() {
	app := kingpin.New(os.Args[0], usage)
	app.Version("0.3")

	verbose = app.Flag("verbose", "Enable verbose output").Short('v').Bool()
	uname = app.Flag("user", "Use current $USER for superuser").Short('u').Bool()

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		app.UsageErrorf(os.Stderr, "%s", err)
	}
}

func main() {
	var username = "postgres"
	if *uname {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("Unable to determine current user!")
		}
		username = u.Username
	}

	pg := &pgembed.PgEmbed{Superuser: username}
	if *verbose {
		pg.Out = os.Stdout
	}

	err := pg.Start()
	if err != nil {
		log.Fatalf("unable to start postgres: %s", err)
	}

	ctl := make(chan interface{})
	go func() {
		<-ctl
		err := pg.Stop()
		if err != nil {
			log.Fatal("unable to stop postgres: %s", err)
			fmt.Fprintf(os.Stderr, "unable to stop postgres: %s", err)
		}
		close(ctl)
	}()

	fmt.Printf("user\t%s\n", pg.Superuser)
	fmt.Printf("data\t%s\n", pg.DataDir)
	fmt.Printf("port\t%d\n", pg.Port)
	fmt.Printf("pid\t%d\n", pg.Pid())
	if *uname {
		fmt.Printf("repl\tpsql -p %d postgres\n", pg.Port)
	} else {
		fmt.Printf("repl\tpsql -U postgres -p %d %s\n", pg.Port, pg.Superuser)
	}

	// await signal to exit
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	ctl <- (<-c)
	<-ctl
}

const usage = `Starts a PostgreSQL database on a random high port and deletes the database when this process exits (C-c). 

Auth is set to 'trust' (no passwords!), and the default superuser is 'postgres' unless the -u flag is given, in which case the superuser will match the current username.`
