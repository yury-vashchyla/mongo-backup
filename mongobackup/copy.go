/*
** copy.go for copy.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Thu 24 Dec 23:43:24 2015 gaspar_d
** Last update Mon  7 Mar 16:52:44 2016 gaspar_d
 */

package mongobackup

import (
	"os"
	"time"

        "github.com/codeskyblue/go-sh"
)

// Return the total size of the directory in byte
func (e *BackupEnv) GetDirSize(source string) int64 {
	directory, _ := os.Open(source)
	var sum int64 = 0
	defer directory.Close()

	objects, _ := directory.Readdir(-1)
	for _, obj := range objects {
		if obj.IsDir() {
			sum += e.GetDirSize(source + "/" + obj.Name())
		} else {
			stat, _ := os.Stat(source + "/" + obj.Name())
			sum += stat.Size()
		}
	}

	return sum
}

func (e *BackupEnv) GetDestFileName(dest string) string {
	t := time.Now()
        return dest + "/" + e.Options.Prefix + "-" + e.Options.BackupType + "-" + t.Format("20060102") + ".tar.bz2.aes"
}

func (e *BackupEnv) TarDir(source string, dest string) (err error, backedByte int64) {
        destfilename := e.GetDestFileName(dest)
	_, err = sh.Command("mkdir","-p",dest).Command("tar", "-cf", "-", "-j", ".", sh.Dir(source)).Command("openssl", "enc", "-e", "-aes-128-cbc", "-k", e.Options.EncPasswd, "-out", destfilename).Output()
	if err != nil {
		return err, 0
	}
	stat, _ := os.Stat(destfilename)
	return nil, stat.Size()
}

func (e *BackupEnv) UnTar(tarfile string, outputdir string) (err error, backedByte int64) {
	_, err = sh.Command("mkdir","-p",outputdir).Command("openssl", "enc", "-d", "-aes-128-cbc", "-k", e.Options.EncPasswd, "-in", tarfile).Command("tar", "-xf", "-", "-j", "-C", outputdir).Output()
	if err != nil {
		return err, 0
	}
	return nil, e.GetDirSize(outputdir)
}


func (e *BackupEnv) checkIfDirExist(dir string) error {
	_, err := os.Stat(dir)
	return err
}
