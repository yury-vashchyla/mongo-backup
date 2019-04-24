package mongobackup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/codeskyblue/go-sh"
	"github.com/minio/minio-go"
)

// Perform uploading according to the command line options
func (e *BackupEnv) PerformUpload() error {
	baklist := []string{}
	if e.Options.BackupID != "" {
		entry := e.homeval.GetBackupEntry(e.Options.BackupID)
		if entry == nil {
			return fmt.Errorf("Can not find the backup %s", e.Options.BackupID)
		}
		baklist = append(baklist, entry.Dest)
		err := e.UploadtoS3(baklist)
		if err != nil {
			e.CleanupBackupEnv()
			return fmt.Errorf("Error while uploading backup %s (%s)", e.Options.BackupID, err)
		}
	} else {
		entryFull, dumpfile, err := e.DumpOplog()
		if err != nil {
			e.CleanupBackupEnv()
			return err
		}
    baklist = append(baklist, entryFull)
    if dumpfile != "" {
      baklist = append(baklist, dumpfile)
    }		
		err = e.UploadtoS3(baklist)
		if err != nil {
			e.error.Printf("Error while uploading backup (%s)", err)
			e.CleanupBackupEnv()
			os.Exit(1)
		}
		_, err = sh.Command("rm", "-f", dumpfile).Output()
		if err != nil {
			e.info.Printf("remove file: %s failed with error %s", dumpfile, err)
		}
	}

	return nil
}

// Remove & delete a range of backups
func (e *BackupEnv) UploadtoS3(uploadlist []string) error {
	endpoint := e.Options.EndPoint
	accessKeyID := e.Options.AccessKey
	secretAccessKey := e.Options.SecretKey
	useSSL := e.Options.useSSL
	bucketName := e.Options.BucketName

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return err
	}

	for _, uploadfile := range uploadlist {
		// Upload the zip file
		objectName := filepath.Base(uploadfile)
		contentType := "application/zip"

		e.info.Printf("prepear upload %s to s3", uploadfile)
		_, st_err := minioClient.StatObject(bucketName, objectName, minio.StatObjectOptions{})
		if st_err != nil {
			// Upload the zip file with FPutObject
			e.info.Printf("upload %s to s3", uploadfile)
			_, err := minioClient.FPutObject(bucketName, objectName, uploadfile, minio.PutObjectOptions{ContentType: contentType})
			if err != nil {
				return err
			}
			e.info.Printf("file %s has been uploaded to s3", uploadfile)
		} else {
			e.info.Printf("file %s already in s3, will not upload", uploadfile)
		}
	}
	return nil
}

func (e *BackupEnv) DumpOplog() (string, string, error) {
	if e.Options.Output != "" {
		e.info.Printf("Get Lastest Backup to restore, BackupID: %d", (e.homeval.content.Sequence - 1))
		entry := e.homeval.GetBackupEntry(strconv.Itoa(e.homeval.content.Sequence - 1))
		if entry == nil {
			return "", "", fmt.Errorf("Backup %s can not be found", (e.homeval.content.Sequence - 1))
		}
		if entry.Type == "inc" {
			entryFull := e.homeval.GetLastFullBackup(*entry)
			if entryFull == nil {
				return "", "", fmt.Errorf("Error, can not retrieve a valid full backup before incremental backup %s", entry.Id)
			}
			e.info.Printf("Dumping oplog of the required full backup: %s", entryFull.Id)
			err := e.DumpOplogsToDir(entryFull, entry)
			if err != nil {
				return "", "", fmt.Errorf("Restore of %s failed while dumping oplog (%s)", entryFull.Dest, err)
			}
			_, err = sh.Command("cp", filepath.Dir(filepath.Dir(entry.Dest))+"/backup.json", e.Options.Output+"/oplog/").Output()
			if err != nil {
				e.info.Printf("cp backup.json failed: %s failed with error %s", filepath.Dir(filepath.Dir(entry.Dest))+"/backup.json", err)
			}
			e.TarDir(e.Options.Output+"/oplog/", filepath.Dir(entry.Dest), entryFull.Id)
			e.info.Printf("remove directory: %s", e.Options.Output)
			_, err = sh.Command("rm", "-rf", e.Options.Output).Output()
			if err != nil {
				e.info.Printf("remove directory: %s failed with error %s", e.Options.Output, err)
			}
			return entryFull.Dest, e.GetDestFileName(filepath.Dir(entry.Dest), entryFull.Id), nil
		} else if entry.Type == "full" {
      return entry.Dest, "", nil
    }
	}
	return "", "", fmt.Errorf("Invalid configuration")
}
