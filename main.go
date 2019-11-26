package main

import (
	"bytes"
	// "fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

//carteira
type Carteira struct {
	ID     uuid.UUID `json:"id"`
	Stakes []Stake   `json:"stakes"`
}

//stake
type Stake struct {
	ID         uuid.UUID `json:"id"`
	IDCarteira uuid.UUID `json:"idcarteira"`
	Dados      int       `json:"dados"`
	Tokens     int       `json:"tokens"`
	Sent       bool
}

func criaCarteira(inicial bool) (cart Carteira, ok bool) {
	source := rand.NewSource(time.Now().Unix())
	numero := rand.New(source)
	if inicial {
		cart := Carteira{
			ID: uuid.New(),
		}
		return cart, true
	}
	if numero.Intn(100) > 79 {
		cart := Carteira{
			ID: uuid.New(),
		}
		return cart, true
	}
	return Carteira{}, false
}

func (cart Carteira) geraStake(inicial bool) (stk Stake, ok bool) {
	source := rand.NewSource(time.Now().Unix())
	numero := rand.New(source)
	if inicial {
		stk := Stake{
			ID:         uuid.New(),
			IDCarteira: cart.ID,
			Dados:      rand.Intn(1e5),
			Tokens:     rand.Intn(1e4),
			Sent:       false,
		}
		return stk, true
	}
	if numero.Intn(100) > 79 {
		stk := Stake{
			ID:         uuid.New(),
			IDCarteira: cart.ID,
			Dados:      rand.Intn(1e5),
			Tokens:     rand.Intn(1e4),
			Sent:       false,
		}
		return stk, true
	}
	return Stake{}, false
}

func enviaStake(client *http.Client, req *http.Request) (resp *http.Response, err error) {
	resp, err = client.Do(req)
	defer resp.Body.Close()
	return
}

func main() {
	var carteiras []Carteira
	for i := 0; i < 15; i++ {
		cart, _ := criaCarteira(true)
		stk, _ := cart.geraStake(true)
		cart.Stakes = append(cart.Stakes, stk)
		carteiras = append(carteiras, cart)
	}
	spew.Dump(carteiras)
	for {
		cart, ok := criaCarteira(false)

		if ok {
			stk, _ := cart.geraStake(true)
			cart.Stakes = append(cart.Stakes, stk)
			carteiras = append(carteiras, cart)
			spew.Dump(carteiras)
		}

		for indiceCart, cart := range carteiras {
			for indiceStk, stk := range cart.Stakes {
				spew.Dump(stk)
				if !stk.Sent {
					stkJSON := []byte(`{"id":"` + stk.ID.String() +
						`", "idcarteira":"` + stk.IDCarteira.String() +
						`", "dados":` + strconv.Itoa(stk.Dados) +
						`, "tokens":` + strconv.Itoa(stk.Tokens) +
						`}`)
					req, err := http.NewRequest("POST", "http://localhost:9000", bytes.NewBuffer(stkJSON))
					if err != nil {
						log.Fatal(err)
					}
					req.Close = true
					req.Header.Set("Content-Type", "application/json")
					client := &http.Client{Timeout: 0 * time.Second}
					_, err = enviaStake(client, req)
					if err != nil {
						log.Fatal(err)
					}
					stk.Sent = true
					cart.Stakes[indiceStk] = stk
				}
			}
			stk, ok := cart.geraStake(false)
			if ok {
				spew.Dump("Carteira " + strconv.Itoa(indiceCart) + " gerou stake.")
				spew.Dump(stk)
			}
		}
		slp := rand.Intn(20)
		spew.Dump(slp)
		time.Sleep(time.Duration(slp) * time.Second)
	}
}
