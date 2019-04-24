# Mongo-backup
***Incremental backup tool for MongoDB***

## Overview
Mongobackup is an external tool performing full & incremental backup. Backup are stored on the filesystem and compressed using the [lz4 algorithm](https://code.google.com/p/lz4/). Full backup are done by performing a file system copy of the dbPath and partial oplog dump is used for incremental backup.

## Features
  * Full & incremental backup
  * Point in Time restore for incremental & full backup 
  * Backup separation by _tag_ (e.g. daily, monthly, weekly, ...)
  * Partial dump of the oplog. Only the operations after the last known one are dumped
  * Secure backup by using [fsyncLock & fsyncUnlock] (https://docs.mongodb.org/manual/reference/method/db.fsyncLock/)
  * Avoid interacting with the primary node

## Usages
Perform an incremental backup
```
./bin/mongo-backup backup [--username string] [--password string] [--backupdir[--backupdir string] [--tag string] [--nocompress] [--nofsynclock] [--stepdown]
```
Perform a full backup         
```
./bin/mongo-backup backup -backuptype full [--username string] [--password string] [-backupdir string] [--tag string] [--nocompress] [--nofsynclock] [--stepdown]
```
Restore a specific backup
```
./bin/mongo-backup restore --restoredir string --backupid string [--username string] [--password string] [--backupdir string]
```
Perform a point in time restore
```
./bin/mongo-backup restore --restoredir string --pit string [--username string] [--password string] [--backupdir string]
```
Perform a oplog dump
```
./bin/mongo-backup oplogdump --restoredir string [--username string] [--password string] [--backupdir string]
```
Delete a range of backup
```
./bin/mongo-backup delete --tag string --entries string [--backupdir string]
```
Delete a specific backup
```
./bin/mongo-backup delete --backupid string [--backupdir string]
```
List available backups
```
./bin/mongo-backup list [--tag string] [--entries string] [--backupdir string]
```
upload backup to s3
```
./bin/mongo-backup upload --restoredir string --accesskey string --secretkey string --bucketname string [--endpoint string] [--nossl] [--backupid string]
```
## Sample configuration

Scheduling has to be performed using an external tool, e.g. cron
Bellow a sample configuration for a daily backup where a full backup is performed twice a week every Sunday, Wednesday and where we stored a daily backup for the last 7 days and backup oplog every 15mins.
```cron
0 0 * * 0,3              mongo-backup backup --backupdir /backup --backuptype full --tag daily --username bakuser --password bakpass && mongo-backup delete --backupdir /backup --tag daily --entries '7-' && mongo-backup upload --restoredir string --accesskey string --secretkey string --bucketname string
*/15 * * * *             mongo-backup backup --backupdir /backup --tag daily --username bakuser --password bakpass
10 0 * * * 1,2,4,5,6      mongo-backup upload --restoredir /dbbackup/tmp --accesskey akey --secretkey skey --bucketname bname --backupdir /dbbackup
```

## Releases

The project is in version **0.01** and is currently **not yet ready for production**.
If you are interested and want more information or want to participate, feel free to ping me ;)
