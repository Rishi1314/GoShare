package router

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"

	"archive/zip"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

func OpenWebsite(url string) {
	fmt.Println("Opening website: ", url)
	var error error
	switch runtime.GOOS {
	case "darwin":
		error=exec.Command("open", url).Start()
	case "linux":
		error=exec.Command("xdg-open", url).Start()
		if error != nil {
			if _, ok := error.(*exec.Error); ok {
				log.Println("xdg-open not found in PATH, please open the URL manually:", url)
				return
			}
		}
	case "windows":
		error=exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		error=fmt.Errorf("unsupported platform")

	}
	if error != nil {
		log.Fatal(error)
	}

}

func GetIP() string {
	connection,error:=net.Dial("udp", "8.8.8.8:80")
	fmt.Println("In GetIP")

	if error != nil {
		log.Fatal(error)
	}
	defer connection.Close()

	localAddr := connection.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func hashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	} // GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func generateQR() string {
	url := "http://" + GetIP() + ":8080"
	qrCode, _ := qrcode.New(url, qrcode.High)
	bytes, err := qrCode.PNG(256)
	if err != nil {
		panic(err)
	}
	imgBase64Str := base64.StdEncoding.EncodeToString(bytes)
	return imgBase64Str
}

func displayError(c *gin.Context, message string, err error) {
	print(message + " <<<- this error @ this endpoint ->>> ") // to print all erroe at console
	c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"Note": message, "Error": err})
}


func getDir() string {
	path, err := filepath.Abs("sync.io-cache")
	if err != nil {
		panic(err)
	}
	exist, _ := exists(path)
	if exist {
		return path
	} else {
		os.MkdirAll(path, 0777)
		return path
	}
}
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func zipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func getFilesInDir(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
