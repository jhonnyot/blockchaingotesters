package main

import (
	"bytes"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const url = "http://localhost:8080"

func postRequest(dificuldade int) error {
	rand.Seed(time.Now().UnixNano())
	dados := rand.Intn(1e6)
	reqJSON := []byte(`{"Dados":` + strconv.Itoa(dados) + `, "Dificuldade":` + strconv.Itoa(dificuldade) + `}`)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 0 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("Resposta: " + resp.Status)
	//fmt.Println("Corpo: ", string(body))
	return nil
}

func getRequest() {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(resp)
}

func main() {
	for {
		milis := rand.Intn(99)
		secs := rand.Intn(60)
		t := (time.Duration(secs) * time.Second) + (time.Duration(milis) * time.Millisecond)
		// go postRequest()
		time.Sleep(t)
	}

}
