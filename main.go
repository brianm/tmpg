package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
)

var verbose = flag.Bool("v", false, "verbose")
var help = flag.Bool("h", false, "show help")

func main() {
	flag.Parse()

	if *help {
		fmt.Println(`USAGE: tmpg [flags]
  Starts a PostgreSQL database on a random high port and
  deletes the database when this process exits (C-c).

FLAGS
  -v  Verbose output
  -h  Show this help
`)
		return
	}

	data_dir, err := ioutil.TempDir(os.TempDir(), "tmpg.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(data_dir)

	u, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to determine current user!")
	}
	err = initdb(data_dir, u.Username)
	if err != nil {
		log.Fatalf("Unable to initialize data dir: %s", err)
	}

	port := availPort()

	ctl, err := run(data_dir, port)
	if err != nil {
		log.Fatalf("unable to start postgres: %s", err)
	}

	fmt.Printf("data\t%s\n", data_dir)
	fmt.Printf("port\t%d\n", port)
	fmt.Printf("repl\tpsql -p %d postgres\n", port)

	// await signal to exit
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	ctl <- (<-c)
	<-ctl
}

func run(dataDir string, port int) (chan os.Signal, error) {
	cmd, err := exec.LookPath("postgres")
	if err != nil {
		return nil, err
	}

	postgres := exec.Command(cmd,
		"-D", dataDir,
		"-p", strconv.Itoa(port),
		"-i",
		"-F")

	if *verbose {
		postgres.Stdout = os.Stdout
		postgres.Stderr = os.Stderr
	}

	err = postgres.Start()
	if err != nil {
		return nil, err
	}

	ctl := make(chan os.Signal)
	go awaitShutdown(postgres, ctl)

	return ctl, nil
}

func awaitShutdown(postgres *exec.Cmd, ctl chan os.Signal) {
	sig := <-ctl
	err := postgres.Process.Signal(sig)
	if err != nil {
		log.Fatalf("unable to kill postgres: %s", err)
	}
	close(ctl)
}

func initdb(dataDir string, username string) error {
	cmd, err := exec.LookPath("initdb")
	if err != nil {
		return err
	}
	init := exec.Command(cmd,
		"-A", "trust",
		"-U", username,
		"-D", dataDir,
		"-E", "UTF-8")
	if *verbose {
		init.Stdout = os.Stdout
		init.Stderr = os.Stderr
	}
	return init.Run()
}

func availPort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Unable to listen on random high port: %s", err)
	}
	defer l.Close()
	parts := strings.Split(l.Addr().String(), ":")
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		log.Fatalf("Unable to parse port: %s", err)
	}
	return port
}
