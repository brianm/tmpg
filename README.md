# Temporary PostgreSQL

Small tool to create and start an isolated postgres instance on a
random high port.

# To Install

The only way right now is to build it from source. It's in Go, so:

```
go get github.com/brianm/tmpg
```

# Usage
```
USAGE: tmpg [flags]
  Starts a PostgreSQL database on a random high port and
  deletes the database when this process exits (C-c).

  Auth is set to 'trust' (no passwords!), and the default
  superuser is 'postgres' unless the -u flag is given,
  in which case the superuser will match the current
  username.

FLAGS
  -v  Verbose output
  -u  superuser set to current username instead of 'postgres'
  -h  Show this help
```
