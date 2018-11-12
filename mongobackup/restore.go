/*
** restore.go for restore.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 23:33:35 2015 gaspar_d
** Last update Mon  7 Mar 16:53:59 2016 gaspar_d
*/

package mongobackup

import (
  "os"
  "strings"
  "strconv"
  "time"
  "path/filepath"
  "github.com/codeskyblue/go-sh"
)

// perform the restore & dump the oplog if required
// oplog is not automatically replayed (futur impprovment?)
// to restore incremental backup or point in time, mongorestore
// has to be used
func (e *BackupEnv) PerformRestore() {
  var (
    entry *BackupEntry
  )

  if e.Options.Output != "" {
    if e.Options.BackupID != "" {
      entry = e.homeval.GetBackupEntry(e.Options.BackupID)
      if entry == nil {
        e.error.Printf("Backup %s can not be found", e.Options.BackupID)
        e.CleanupBackupEnv()
        os.Exit(1)
      }
    } else if e.Options.Pit != "" {
      pit   := e.Options.Pit
      index := strings.Index(pit, ":")
      if index != -1 {
        pit = pit[:index]
      }

      i, err := strconv.ParseInt(pit, 10, 64)
      if err != nil {
        e.error.Printf("Invalid point in time value: %s (%s)", e.Options.Pit, err)
        e.CleanupBackupEnv()
        os.Exit(1)
      }
      ts := time.Unix(i, 0)

      entry = e.homeval.GetLastEntryAfter(ts)
      if entry == nil {
        e.error.Printf("A plan to restore to the date %s can not be found", ts)
        e.CleanupBackupEnv()
        os.Exit(1)
      }

      err = e.homeval.CheckIncrementalConsistency(entry)
      if err != nil {
        e.error.Printf("Plan to restore the date %s is inconsistent (%s)", e.Options.Pit, err)
        e.CleanupBackupEnv()
        os.Exit(1)
      }
    } else {
      e.info.Printf("Get Lastest Backup to restore, BackupID: %s", (e.homeval.content.Sequence-1))
      entry = e.homeval.GetBackupEntry(strconv.Itoa(e.homeval.content.Sequence-1))
      if entry == nil {
        e.error.Printf("Backup %s can not be found", e.Options.BackupID)
        e.CleanupBackupEnv()
        os.Exit(1)
      }

    }
    e. performFullRestore(entry)
  } else {
    e.error.Printf("Invalid configuration")
    e.CleanupBackupEnv()
    os.Exit(1)
  }
}

// perform the restore & dump of the oplog
func (e *BackupEnv) performFullRestore(entry *BackupEntry) {
  var (
    entryFull *BackupEntry
  )
//  err = e.checkIfDirExist(e.Options.Output)
//  e.info.Printf("Performing a restore of backup %s", entry.Id);
//  if err != nil {
//    e.error.Printf("Can not access directory %s, cowardly failling (%s)", e.Options.Output, err)
//    e.CleanupBackupEnv()
//    os.Exit(1)
//  }

  if entry.Type == "inc" {
    entryFull = e.homeval.GetLastFullBackup(*entry)
    if entryFull == nil {
      e.error.Printf("Error, can not retrieve a valid full backup before incremental backup %s", entry.Id)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
    e.info.Printf("Restoration of backup %s is needed first", entryFull.Id)
  } else {
    entryFull = entry
  }

  if !e.Options.DumpOplog {
    err, restored := e.UnTar(entryFull.Dest, e.Options.Output)

    if err != nil {
      e.error.Printf("Restore of %s failed (%s)", entryFull.Dest, err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
    e.info.Printf("Sucessful restoration, %fGB has been restored to %s", float32(restored) / (1024*1024*1024), e.Options.Output)
  }

  if entry.Type == "inc" {
    e.info.Printf("Dumping oplog of the required full backup %s", entryFull.Id)
    err := e.DumpOplogsToDir(entryFull, entry)
    if err != nil {
      e.error.Printf("Restore of %s failed while dumping oplog (%s)", entryFull.Dest,  err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
    if e.Options.DumpOplog {
      e.TarDir(e.Options.Output + "/oplog/", filepath.Dir(entry.Dest)) 
      e.info.Printf("remove directory: %s", e.Options.Output)
      _, err = sh.Command("rm -rf "+e.Options.Output + "/oplog").Command("rmdir "+e.Options.Output).Output()
      if err != nil {
         e.info.Printf("remove directory: %s failed with error %s", e.Options.Output, err)
      }
      e.info.Printf("oplog dump finish with filename: %s", e.GetDestFileName(filepath.Dir(entry.Dest)))
    } else {
      message := "Success. To replay the oplog, start mongod and execute: "
      if e.Options.Pit == "" {
        message += "`mongorestore --oplogReplay " + e.Options.Output + "/oplog/`"
      } else {
        message += "`mongorestore --oplogReplay --oplogLimit " +  e.Options.Pit + " " + e.Options.Output + "/oplog/`"
      }
      e.info.Printf(message)
    }
  }
}
