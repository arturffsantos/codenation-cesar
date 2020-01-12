package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type CesarMessage struct {
	NumeroCasas         int    `json:"numero_casas"`
	Token               string `json:"token"`
	Cifrado             string `json:"cifrado"`
	Decifrado           string `json:"decifrado"`
	ResumoCriptografico string `json:"resumo_criptografico"`
}

func (m CesarMessage) String() string {
	return fmt.Sprintf("NumeroCasas: %v, Token: %v, Cifrado: %v, Decifrado: %v, ResumoCriptografico: %v  ", m.NumeroCasas, m.Token, m.Cifrado, m.Decifrado, m.ResumoCriptografico)
}

func (message *CesarMessage) Decode() {
	i := 0
	var charOriginal int
	var mensagemOriginal string
	var deslocamento = 26 - (message.NumeroCasas % 26)
	cifrado := strings.ToLower(message.Cifrado)
	for _, char := range cifrado {
		if char >= 97 && char <= 122 {
			charOriginal = (int(char)-97+deslocamento)%26 + 97
			mensagemOriginal += string(charOriginal)
		} else {
			mensagemOriginal += string(char)
		}

		i++
	}
	message.Decifrado = mensagemOriginal
	return
}

func (message *CesarMessage) GenerateSha1() {
	data := []byte(message.Decifrado)
	message.ResumoCriptografico = fmt.Sprintf("%x", sha1.Sum(data))
}

func GetMessage(token string) CesarMessage {
	getURL := "https://api.codenation.dev/v1/challenge/dev-ps/generate-data?token=" + token
	resp, err := http.Get(getURL)
	if err != nil {
		fmt.Printf("error get: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	var message CesarMessage
	err = json.Unmarshal(body, &message)
	if err != nil {
		fmt.Printf("erro decode %v", err)
	}

	return message
}

func SendFile(fileName string, token string) {
	fileDir, _ := os.Getwd()
	filePath := path.Join(fileDir, fileName)

	file, _ := os.Open(filePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("answer", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	postURL := "https://api.codenation.dev/v1/challenge/dev-ps/submit-solution?token="+ token
	r, _ := http.NewRequest("POST", postURL, body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	client.Do(r)
}

func SaveFile(message CesarMessage) {
	jsonBody, _ := json.Marshal(message)
	err := ioutil.WriteFile("answer.json", jsonBody, 0644)
	if err != nil {
		fmt.Printf("erro writefile %v", err)
	}
}

func main() {
	fmt.Printf("Cifra de César\n")

	token := os.Args[1];

	message := GetMessage(token)
	fmt.Print(message)

	//decodificando
	message.Decode()
	fmt.Println("\nafter decode")
	fmt.Print(message)

	//gerando o sha1
	message.GenerateSha1()
	fmt.Println("\nafter SHA1")
	fmt.Print(message)

	SaveFile(message)

	//enviado o arquivo
	SendFile("answer.json")
}
