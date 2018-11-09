/*
** options.go for options.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:28:29 2015 gaspar_d
** Last update Mon  7 Mar 16:53:55 2016 gaspar_d
 */

package mongobackup

import (
	"fmt"
	"os"

	"github.com/pborman/getopt"
)

const (
	OpBackup  = 0
	OpRestore = 1
	OpList    = 4
	OpDelete  = 8

	DefaultTag = "daily"
	DefaultDir  = "mongobak"
)

// abstract structure standing for command line options
type Options struct {
	// general options
	Operation int
	Directory string
	Tag      string
	Stepdown  bool
	Position  string
	Debug     bool
	// backup options
	Fsynclock   bool
	BackupType string
	Compress    bool
        EncPasswd   string
        Prefix      string
	// mongo options
	Mongohost string
	Mongouser string
	Mongopwd  string
	// restore options
	Output   string
	Pit      string
	BackupID string
}

// parse the command line and create the Options struct
func ParseOptions() Options {
	var (
		lineOption Options
		set        *getopt.Set
	)

	set = getopt.New()
	pwd, _ := os.Getwd()

	optDirectory := set.StringLong("backupdir", 'b', pwd+"/"+DefaultDir, "")
	optTag := set.StringLong("tag", 't', DefaultTag, "")
	optStepdown := set.BoolLong("stepdown", 0, "")
	optNoFsyncLock := set.BoolLong("nofsynclock", 0, "")
	optNoCompress := set.BoolLong("nocompress", 0, "")
	optBackupType := set.StringLong("backuptype", 0, "inc", "")
	optHelp := set.BoolLong("help", 'h', "")
	optDebug := set.BoolLong("debug", 'd', "")

	optMongo := set.StringLong("host", 0, "localhost:27017", "")
	optMongoUser := set.StringLong("username", 'u', "", "")
	optMongoPwd := set.StringLong("password", 'p', "", "")

	optEncPasswd := set.StringLong("encpasswd", 'e', "d0cker", "")
	optPrefix := set.StringLong("prefix", 0, "mongobak", "")
	optPitTime := set.StringLong("pit", 0, "", "")
	optBackupID := set.StringLong("backupid", 0, "", "")
	optOutput := set.StringLong("restoredir", 'r', "", "")

	optPosition := set.StringLong("entries", 0, "", "")

	set.SetParameters("backup|restore|list")

	err := set.Getopt(os.Args[1:], nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		set.PrintUsage(os.Stdout)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		PrintHelp()
		os.Exit(1)
	} else if os.Args[1] == "backup" {
		lineOption.Operation = OpBackup
	} else if os.Args[1] == "restore" {
		lineOption.Operation = OpRestore
	} else if os.Args[1] == "list" {
		lineOption.Operation = OpList
	} else if os.Args[1] == "delete" {
		lineOption.Operation = OpDelete
	} else if os.Args[1] == "help" || (*optHelp) {
		PrintHelp()
		os.Exit(0)
	} else {
		PrintHelp()
		os.Exit(1)
	}

	lineOption.Stepdown = *optStepdown
	lineOption.Fsynclock = !*optNoFsyncLock
	lineOption.BackupType = *optBackupType
	lineOption.Directory = *optDirectory
	lineOption.Compress = !*optNoCompress
	lineOption.Debug = *optDebug

	lineOption.Mongohost = *optMongo
	lineOption.Mongouser = *optMongoUser
	lineOption.Mongopwd = *optMongoPwd
	lineOption.EncPasswd = *optEncPasswd
	lineOption.Prefix = *optPrefix
	lineOption.Tag = *optTag
	lineOption.Pit = *optPitTime
	lineOption.BackupID = *optBackupID
	lineOption.Output = *optOutput
	lineOption.Position = *optPosition

	if !validateOptions(lineOption) {
		getopt.Usage()
		os.Exit(1)
	}

	return lineOption
}

// validate the option to see if there is
// any incoherence (TODO)
func validateOptions(o Options) bool {
	return true
}

func PrintHelp() {
	var helpMessage []string

	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "-b", "--backupdir=string", "directory to save backup"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "-k", "--tag=string", "metadata associated to the backup"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--stepdown", "rs.stepDown() if this is the primary node"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--nofsynclock", "Avoid using fsyncLock() and fsyncUnlock()"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--nocompress", "disable compression for backup & restore"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--backuptype", "backup type [inc, full]"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--host=string", "mongo hostname"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "-u", "--username=string", "mongo username"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "-p", "--password=string", "mongo password"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--prefix=string", "backup file name prefix"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "-e", "--encpasswd=string", "encode password"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--pit=string", "point in time recovery (using oplog format: unixtimetamp:opcount)"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--backupid=string", "to restore a specific backup"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "-r", "--restoredir=string", "directory to restore"))
	helpMessage = append(helpMessage, fmt.Sprintf("%-5s %-20s %s", "", "--entries=string", "criteria string (format number[+-])"))

	fmt.Printf("\nUsage:\n\n    %s command options\n", os.Args[0])

	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("\n")
	fmt.Printf("    %-35s %s %s\n", "perform an incremental backup", os.Args[0], "backup [--tag string] [--nocompress] [--nofsynclock] [--stepdown]")
	fmt.Printf("    %-35s %s %s\n", "perform a full backup", os.Args[0], "backup -backuptype full [--tag string] [--nocompress] [--nofsynclock] [--stepdown]")
	fmt.Printf("    %-35s %s %s\n", "restore a specific backup", os.Args[0], "restore --restoredir string --backupid string")
	fmt.Printf("    %-35s %s %s\n", "perform a point in time restore", os.Args[0], "restore --restoredir string --pit string")
	fmt.Printf("    %-35s %s %s\n", "delete a range of backup", os.Args[0], "delete --tag string --entries string")
	fmt.Printf("    %-35s %s %s\n", "delete a specific backup", os.Args[0], "delete --backupid string")
	fmt.Printf("    %-35s %s %s\n", "list available backups", os.Args[0], "list [--tag string] [--entries string]")
	fmt.Printf("\n")
	fmt.Printf("Options:\n")
	fmt.Printf("\n")

	for _, help := range helpMessage {
		fmt.Print("    ")
		fmt.Print(help)
		fmt.Print("\n")
	}
}
