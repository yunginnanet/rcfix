package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

const defaultTemplate = `#!/bin/sh
# /etc/init.d/{{.Description}}
#
# {{.Description}}
#
### BEGIN INIT INFO
# Provides:          {{.Name}}
# Required-Start:    $local_fs $network $named $time $syslog
# Required-Stop:     $local_fs $network $named $time $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Description:       {{.Description}}
### END INIT INFO

_DONE=".spinstop"
_CLR=1
cln() {
	tput rc
	tput cnorm
	return 1
}
trap cln INT QUIT TERM
spinner() {
	if [ -t 1 ]; then
		_SYM=". o O o"
		_ppid=$(ps -p "$$" -o ppid=)
		while :; do
			tput civis
			for c in $_SYM; do
				tput sc
				env printf "$c"
				tput rc
				if [ -f "$_DONE" ]; then
					if [ $_CLR -eq 1 ]; then
						tput el
					fi
					rm "$_DONE"
					break 2
				fi
				env sleep .2
				if [ -n "$_ppid" ]; then
					_pup=$(ps --no-headers "$_ppid")
					if [ -z "$_pup" ]; then
						break 2
					fi
				fi
			done
		done
		tput cnorm
		return 0

	fi
}

PIDFILE="/var/run/{{.Name}}.pid"
LOGFILE="/var/log/{{.Name}}.log"
SCRIPT="{{.ExecStart}}"
RUNAS="{{.User}}"

start() {
  if [ -f "$PIDFILE" ] && [ -s "$PIDFILE" ] && kill -0 $(cat $PIDFILE); then
    echo 'Service already running, PID: $(cat PIDFILE)' >&2
    return 1
  fi
  spinner &
  echo 'Starting service…' >&2
  local CMD="$SCRIPT &> \"$LOGFILE\" & echo \$!"
  su -c "$CMD" $RUNAS > "$PIDFILE"

  sleep 2
  PID=$(cat $PIDFILE)
    if pgrep -u $RUNAS -f $NAME > /dev/null
    then
	  touch $_DONE
      echo "$NAME is now running, the PID is $PID"
    else
      touch $_DONE
      echo ''
      echo "Error! Could not start $NAME, check $LOGFILE for more information."
    fi
}

stop() {
  if [ ! -f "$PIDFILE" ] || ! kill -0 $(cat "$PIDFILE"); then
    echo 'Service not running' >&2
    return 1
  fi
  echo 'Stopping service…' >&2
  spinner &
  kill -15 $(cat "$PIDFILE") && rm -f "$PIDFILE"
  touch $_DONE
  echo 'Service stopped' >&2
}

status() {
    printf "%-50s" "Checking {{.Name}}..."
	spinner &
    if [ -f $PIDFILE ] && [ -s $PIDFILE ]; then
        PID=$(cat $PIDFILE)
            if [ -z "$(ps axf | grep ${PID} | grep -v grep)" ]; then
				touch $_DONE
                printf "%s\n" "The process appears to be dead but pidfile still exists"
            else    
				touch $_DONE
                echo "Running, the PID is $PID"
            fi
    else
		touch $_DONE
        printf "%s\n" "Service not running"
    fi
}


case "$1" in
	start)
		start
		;;
	stop)
		stop
		;;
{{if .HasReload}}
	reload)
		echo "Reloading {{.Name}}"
		spinner &
		{{.ExecReload}}
		touch $_DONE
		;;
{{end}}
	status)
		status()
		;;
	restart)
		$0 stop
		$0 start
		;;
	*)
		echo "Usage: $0 {start|stop|reload|status|restart}"
		exit 1
		;;
esac

exit 0
`

type BrokenService struct {
	Name             string
	Description      string
	Documentation    string
	ExecStart        string
	ExecStop         string
	ExecReload       string
	PartOf           string
	Service          string
	Restart          string
	RuntimeDirectory string
	RemainAfterExit  bool
	HasReload        bool
	Requires         string
	Wants            string
	After            string
	Before           string
	User             string
	Group            string
}

func (bs *BrokenService) parseLine(line string) {
	var index = map[string]*string{
		"Description": &bs.Description, "Documentation": &bs.Documentation,
		"ExecStart": &bs.ExecStart, "ExecStop": &bs.ExecStop,
		"ExecReload": &bs.ExecReload, "Restart": &bs.Restart,
		"Requires": &bs.Requires, "Wants": &bs.Wants,
		"After": &bs.After, "Before": &bs.Before,
		"RuntimeDirectory": &bs.RuntimeDirectory, "#RuntimeDirectory": &bs.RuntimeDirectory,
	}
	for key, value := range index {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			*value = strings.TrimPrefix(line, key+"=")
			return
		}
	}
}

type ServiceFixer struct {
	Name     string
	Template *template.Template
	Corpse   BrokenService
	buf      []byte
}

func (s *ServiceFixer) Write(p []byte) (int, error) {
	reader, writer := io.Pipe()
	type writeRes struct {
		n   int
		err error
	}
	resChan := make(chan writeRes)
	go func() {
		defer func() {
			_ = writer.Close()
		}()
		n, err := writer.Write(p)
		resChan <- writeRes{n, err}
	}()
	sad, sadErr := ReadBrokenService(reader)
	if sadErr != nil {
		return 0, sadErr
	}
	s.Corpse = sad
	s.Template = template.Must(template.New("sysvinit").Parse(defaultTemplate))
	s.buf = make([]byte, 0)

	res := <-resChan
	return res.n, res.err
}

func (s *ServiceFixer) Fix() error {
	var output strings.Builder
	s.Corpse.Name = s.Name
	if s.Template == nil {
		s.Template = template.Must(template.New("sysvinit").Parse(defaultTemplate))
	}
	if s.Corpse.ExecStart == "" {
		return errors.New("ExecStart is required")
	}
	if s.Corpse.Name == "" {
		return errors.New("name is required, try using the -n flag")
	}
	err := s.Template.Execute(&output, s.Corpse)
	if err != nil {
		return err
	}
	s.buf = []byte(output.String())
	return nil
}

func (s *ServiceFixer) Read(p []byte) (int, error) {
	if len(s.buf) == 0 {
		return 0, io.EOF
	}
	if err := s.Fix(); err != nil {
		return 0, err
	}
	n := copy(p, s.buf)
	s.buf = s.buf[n:]
	return n, nil
}

func ReadBrokenServiceFile(filepath string) (BrokenService, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return BrokenService{}, err
	}
	defer func() {
		_ = file.Close()
	}()
	return ReadBrokenService(file)
}

func ReadBrokenService(r io.Reader) (BrokenService, error) {
	verySad := &BrokenService{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if scanner.Err() != nil && !errors.Is(scanner.Err(), io.EOF) {
			return *verySad, scanner.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		verySad.parseLine(line)
	}

	if (verySad.Name == "" || strings.Contains(verySad.Name, " ")) && verySad.RuntimeDirectory != "" {
		verySad.Name = verySad.RuntimeDirectory
	}

	if verySad.ExecReload != "" {
		verySad.HasReload = true
	}

	return *verySad, nil
}

func main() {
	var (
		err  error
		src  io.Reader
		dst  io.Writer
		name string
	)

	fubar := "name flag present but no name provided"

	for i, arg := range os.Args {
		switch arg {
		case "-n", "--name":
			if len(os.Args) < i+1 {
				log.Fatal(fubar)
			}
			name = os.Args[i+1]
			os.Args = slices.Delete(os.Args, i, i+1)
			if strings.Contains(arg, "=") {
				name = strings.Split(arg, "=")[1]
				if name == "" {
					log.Fatal(fubar)
				}
			}
		default:
			continue
		}
	}

	if len(os.Args) > 1 {
		if src, err = os.Open(os.Args[1]); err != nil {
			log.Fatal(err)
		}
		name = strings.TrimSuffix(filepath.Base(os.Args[1]), filepath.Ext(os.Args[1]))
	}
	if len(os.Args) > 2 {
		if dst, err = os.Create(os.Args[2]); err != nil {
			log.Fatal(err)
		}
	}
	if src == nil {
		src = os.Stdin
	}
	if dst == nil {
		dst = os.Stdout
	}
	sad, sadErr := ReadBrokenService(src)
	if sadErr != nil {
		log.Fatal(sadErr)
	}
	if name == "" {
		name = sad.Description
		if strings.Contains(name, " ") && sad.RuntimeDirectory != "" {
			name = filepath.Base(sad.RuntimeDirectory)
		}
	}
	sad.Name = name
	s := &ServiceFixer{Name: name, Corpse: sad}
	_, err = io.Copy(s, src)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Fix()
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(dst, s)
	if err != nil {
		log.Fatal(err)
	}
}
