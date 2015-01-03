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

func main() {
	flag.Parse()

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

	fmt.Printf("psql -p %d template1\n", port)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	ctl <- sig
	<-ctl
	log.Println("exiting!")
}

func run(dataDir string, port int) (chan interface{}, error) {
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
	postgres.Start()

	ctl := make(chan interface{})

	go func() {
		<-ctl
		err := postgres.Process.Kill()
		if err != nil {
			log.Fatalf("unable to kill postgres: %s", err)
		}
		close(ctl)
	}()

	return ctl, nil
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
