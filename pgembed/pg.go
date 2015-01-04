package pgembed

import (
	"gopkg.in/jackc/pgx.v2"

	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PgEmbed struct {
	Port      uint16
	DataDir   string
	Superuser string
	Out       io.Writer

	postgres *exec.Cmd
	starter  sync.Once
	stopper  sync.Once
}

func (pg *PgEmbed) Start() error {
	var err error = nil
	pg.starter.Do(func() {
		if pg.Port == 0 {
			pg.Port = AvailPort()
		}

		if pg.DataDir == "" {
			pg.DataDir, err = ioutil.TempDir(os.TempDir(), "tmpg.")
			if err != nil {
				return
			}
		}

		if pg.Superuser == "" {
			pg.Superuser = "postgres"
		}

		var cmd string

		cmd, err = exec.LookPath("initdb")
		if err != nil {
			return
		}

		init := exec.Command(cmd,
			"-A", "trust",
			"-U", pg.Superuser,
			"-D", pg.DataDir,
			"-E", "UTF-8")
		init.Stdout = pg.Out
		init.Stderr = pg.Out
		err = init.Run()
		if err != nil {
			return
		}

		cmd, err = exec.LookPath("postgres")
		if err != nil {
			return
		}

		pg.postgres = exec.Command(cmd,
			"-D", pg.DataDir,
			"-p", strconv.Itoa(int(pg.Port)),
			"-i",
			"-F")
		pg.postgres.Stdout = pg.Out
		pg.postgres.Stderr = pg.Out

		err = pg.postgres.Start()
		if err != nil {
			return
		}

		// now try to connect until it is up!
		for {
			time.Sleep(5 * time.Millisecond)
			if tryConnect(pg.Superuser, pg.Port) {
				return
			}
		}
	})

	return err
}

func tryConnect(user string, port uint16) bool {
	cfg := pgx.ConnConfig{
		Host:     "localhost",
		Port:     port,
		Database: "postgres",
		User:     user,
	}

	c, err := pgx.Connect(cfg)
	if err != nil {
		return false
	}
	defer c.Close()
	return c.IsAlive()
}

func (pg *PgEmbed) Stop() error {
	if err := pg.Start(); err != nil {
		return err
	}

	var err error
	pg.stopper.Do(func() {
		err = pg.postgres.Process.Kill()
		if err != nil {
			return
		}
		_, err = pg.postgres.Process.Wait()
	})
	return err
}

func (pg *PgEmbed) Pid() int {
	return pg.postgres.Process.Pid
}

func AvailPort() uint16 {
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
	return uint16(port)
}
