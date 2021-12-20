package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/scorredoira/email"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

type EmailConfig struct {
	Username string
	Password string
	Host     string
	Port     int
}

var workers = make(map[string]string)

func init() {
	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	err := license.SetMeteredKey("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	if err != nil {
		panic(err)
	}
}

func DirectoryChk(folderName string) error {
	files, _ := os.ReadDir(".")
	for _, file := range files {
		fmt.Println(file.Name())
		if file.Name() == folderName {
			if file.IsDir() {
				fmt.Println("이것은 디렉터리")
				return nil
			}
		}
	}
	return errors.New("\"input\" 디렉터리가 존재 하지 않음")
}

func main() {
	t := time.Now()
	var folderInput string = "input"
	var folderOutput string = "output"

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = DirectoryChk(folderInput)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	files, err := os.ReadDir(folderInput)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pdf") {
			inputPath := fmt.Sprintf("%s/%s", folderInput, file.Name())
			cmd := exec.Command("split.exe", "split", inputPath, "output")
			output, err := cmd.Output()

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(string(output))
			}
		}
	}

	smtpHost := "smtp.worksmobile.com" // change to your SMTP provider address
	smtpPort := xxx                    // change to your SMPT provider port number
	smtpPass := os.Getenv("smtpPass")  // change here
	smtpUser := os.Getenv("smtpUser")

	emailConf := &EmailConfig{smtpUser, smtpPass, smtpHost, smtpPort}

	emailauth := smtp.PlainAuth("", emailConf.Username, emailConf.Password, emailConf.Host)

	employee, err := ReadCSV()
	if err != nil {
		fmt.Println("error: ", err)
		panic(err)
	}

	fmt.Println("CheckDir() 시작")
	fileList := CheckDir(folderOutput)
	fmt.Println("CheckDir() 끝")
	fmt.Println("fileList: ", fileList)

	for name, path := range employee {
		fmt.Println("workers 메일 보내기 시작!")
		fmt.Println(name, path)
		message := fmt.Sprintf(`
	%s님
	지난 한달 간 수고 많으셨습니다.
	​
	급여명세서를 별첨과 같이 송부합니다.
	(연봉에 대한 정보는 보안 사항이므로, 본인에 대한 사실 외 타인의 연봉 정보에 대한 문의나 확인을 하는 것은 절대로 삼가하여 주시기 바랍니다.)
	​
	개인의 권리 및 의무에 관련하여 보충 설명이나 상세한 이해가 더 필요한 부분은 언제든지 편한 마음으로 연락주시기 바랍니다.
	임의에 의한 판단이나 선입견으로 인하여 불이익을 당하는 경우가 없도록 해 주세요.
	​
	이상입니다.
	감사합니다.`, name)
		subject := fmt.Sprintf("[XXXX회사/급여명세서]%d년 %d월분(월급여)", t.Year(), t.Month())

		emailContent := email.NewMessage(subject, message)
		emailContent.From = mail.Address{Name: "", Address: smtpUser}

		for _, filename := range fileList {
			_, file := filepath.Split(filename)
			if filename == "" || !strings.Contains(file, name) {
				fmt.Println(file, name, "파일명과 이름 불일치")
				continue
			}
			fmt.Println("filename: ", file, "workers name: ", name)
			fmt.Println("이름 일치")
			err := emailContent.Attach(filename)
			if err != nil {
				fmt.Println(err)
			}
			if addr, ok := employee[name]; ok {
				emailContent.To = []string{addr}
				err := email.Send(smtpHost+":"+strconv.Itoa(emailConf.Port), //convert port number from int to string
					emailauth,
					emailContent)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
			}

		}

	}

}

func CheckDir(folderName string) []string {
	folders, _ := os.ReadDir(folderName)
	fmt.Println(folders)
	len := 0

	for _, f := range folders {
		if strings.HasSuffix(f.Name(), ".pdf") {
			len++
		}
	}
	fmt.Println("len: ", len)

	fileList := make([]string, len)
	for _, f := range folders {
		fmt.Println(f.Name())
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		path := "output/" + f.Name()
		fmt.Println(path)

		arr, err := outputPdfText(path)
		fileList = append(fileList, arr)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("arr: ", arr)
	}
	return fileList
}

func outputPdfText(inputPath string) (string, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return "", err
	}

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return "", err
	}

	// numPages, err := pdfReader.GetNumPages()
	// if err != nil {
	// 	return "", err
	// }
	fmt.Printf("--------------------\n")
	fmt.Printf("PDF to text extraction:\n")
	fmt.Printf("--------------------\n")
	var arr string

	pageNum := 1

	page, err := pdfReader.GetPage(pageNum)

	if err != nil {
		return "", err
	}

	ex, err := extractor.New(page)
	if err != nil {
		return "", err
	}

	text, err := ex.ExtractText()

	if err != nil {
		return "", err
	}

	fmt.Println("------------------------------")
	fmt.Printf("Page %d:\n", pageNum)
	fmt.Printf("\"%s\"\n", text)
	fmt.Println("------------------------------")
	if strings.Contains(text, "사원명:") {
		f.Close()
		t := time.Now()
		fmt.Println("사원명 단어 포함 OK")
		idx := strings.Index(text, "사원명:")
		arr = text[idx+11 : idx+20]
		workers[arr] = fmt.Sprintf("output/홍익과학기술_%d년 %d월_급여명세서_%s.pdf", t.Year(), t.Month(), arr)

		err := os.Rename(inputPath, workers[arr])
		if err != nil {
			fmt.Println(err)
		}
		return workers[arr], err
	}

	f.Close()
	panic(errors.New("사원명 포함되있지 않음, Panic"))
}

func ReadCSV() (map[string]string, error) {
	employee := make(map[string]string)
	file, err := os.Open("급여작업_자동화.csv")
	if err != nil {
		log.Println(err)
		return employee, err
	}

	r := csv.NewReader(bufio.NewReader(file))
	// r.Comma = ';'
	// r.Comment = '-'
	// r.LazyQuotes = true
	fmt.Println(r)

	record, _ := r.Read()
	fmt.Printf("record[3]: %s record[4]: %s\n", record[3], record[4])
	record, _ = r.Read()
	fmt.Printf("record[3]: %s record[4]: %s\n", record[3], record[4])

	for {
		record, err := r.Read()
		if err == io.EOF {
			fmt.Println("io.EOF! => break")
			break
		} else if err != nil {
			fmt.Println("error: ", err)
			return employee, err
		}
		if record[3] == "" || record[4] == "" {
			fmt.Println("공백입니다.")
			continue
		}
		employee[record[3]] = record[4]

	}
	file.Close()
	return employee, nil
}
