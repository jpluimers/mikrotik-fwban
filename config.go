package main

import (
	"flag"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/gcfg.v1"
)

type ConfigMikrotik struct {
	Disabled  bool
	Address   string
	User      string
	Passwd    string
	BanList   string
	Whitelist []string
	Blacklist []string
}

type Config struct {
	Settings struct {
		BlockTime  Duration
		AutoDelete bool
		Verbose    bool
		Port       uint16
	}
	Mikrotik map[string]*ConfigMikrotik
}

var (
	progname = path.Base(os.Args[0])

	cleartime  = flag.Duration("cleartime", 7*24*time.Hour, "Set the live time for dynamicly managed entries.")
	filename   = flag.String("filename", "/etc/mikrotik-fwban.cfg", "Path of the configuration file to read.")
	syslogport = flag.Uint("syslogport", 10514, "UDP port we listen on for syslog formatted messages.")
	autodelete = flag.Bool("autodelete", true, "Autodelete entries when they expire. Aka, don't trust Mikrotik to do it for us.")
	verbose    = flag.Bool("verbose", false, "Be more verbose in our logging.")
	debug      = flag.Bool("debug", false, "Be absolutely staggering in our logging.")

	cfg Config
)

func hasFlag(s string) bool {
	var res bool
	flag.Visit(func(f *flag.Flag) {
		if s == f.Name {
			res = true
		}
	})
	return res
}

func configParse() {
	flag.Parse()

	err := gcfg.ReadFileInto(&cfg, *filename)
	if err != nil {
		log.Fatal(err)
	}
	// Flags override the config file
	if hasFlag("cleartime") {
		cfg.Settings.BlockTime = Duration(*cleartime)
	}
	if hasFlag("activedelete") {
		cfg.Settings.AutoDelete = *autodelete
	}
	if hasFlag("verbose") {
		cfg.Settings.Verbose = *verbose
	}
	if hasFlag("syslogport") {
		cfg.Settings.Port = uint16(*syslogport)
	}
	if cfg.Settings.BlockTime == 0 {
		log.Fatal("Blocktime needs to be non-zero.")
	}

	for _, v := range cfg.Mikrotik {
		if v.Disabled {
			continue
		}
		// Add port 8728 if it was not included
		_, _, err := net.SplitHostPort(v.Address)
		if err != nil {
			// For anything else than missing port, bail.
			if !strings.HasPrefix(err.Error(), "missing port in address") {
				continue
			}
			v.Address = net.JoinHostPort(v.Address, "8728")
		}
		// set default managed addresslist name
		if v.BanList == "" {
			v.BanList = "blacklist"
		}
	}
}