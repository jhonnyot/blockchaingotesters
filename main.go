package main

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	// "fmt"

	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
)

//carteira
type Carteira struct {
	ID       uuid.UUID
	Stakes   []Stake
	Currency int
}

//stake
type Stake struct {
	ID                uuid.UUID
	IDCarteiraOrigem  uuid.UUID
	IDCarteiraDestino uuid.UUID
	Currency          int
	Sent              bool
}

type Bloco struct {
	Indice    int
	Timestamp string
	Dados     []Stake
	Hash      string
	HashAnt   string
	Validador string
}

var Blockchain []Bloco
var tempBlocos []Bloco
var filaDeBlocos []Bloco
var stakes []Stake
var validadores = make(map[string]int)
var anunciador = make(chan string)

//sincronizador; garante não-concorrência nas adições de blocos
var mutex = &sync.Mutex{}

//calculador de hashes
func calculaHash(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashFinal := hasher.Sum(nil)
	return hex.EncodeToString(hashFinal)
}

func calculaHashBloco(bloco Bloco) string {
	dadosStr := 0
	for _, stk := range bloco.Dados {
		dadosStr += stk.Dados
	}
	totalDados := string(bloco.Indice) + bloco.Timestamp + string(bloco.Dados) + bloco.HashAnt
	return calculaHash(totalDados)
}

func (cart Carteira) geraBloco(blocoAnterior Bloco) (Bloco, error) {
	var novoBloco Bloco

	t := time.Now()

	novoBloco.Indice = blocoAnterior.Indice + 1
	novoBloco.Timestamp = t.String()
	novoBloco.Dados = stakes
	novoBloco.HashAnt = blocoAnterior.Hash
	novoBloco.Hash = calculaHashBloco(novoBloco)
	novoBloco.Validador = cart.ID.String()

	return novoBloco, nil
}

func blocoValido(novoBloco, blocoAnterior Bloco) bool {
	if blocoAnterior.Indice+1 != novoBloco.Indice {
		return false
	}
	if blocoAnterior.Hash != novoBloco.HashAnt {
		return false
	}
	if calculaHashBloco(novoBloco) != novoBloco.Hash {
		return false
	}

	return true
}

func escolheValidador() {
	spew.Dump("Validando")
	go func() {
		mutex.Lock()
		for _, candidato := range filaDeBlocos {
			tempBlocos = append(tempBlocos, candidato)
		}
		filaDeBlocos = []Bloco{}
		mutex.Unlock()
	}()
	spew.Dump(filaDeBlocos)
	time.Sleep(3 * time.Second)
	mutex.Lock()
	temp := tempBlocos
	mutex.Unlock()

	loteria := []string{}
	if len(temp) > 0 {
		//percorre o slice de blocos procurando validadores únicos
	EXTERNO:
		for _, bloco := range temp {
			for _, node := range loteria {
				if bloco.Validador == node {
					continue EXTERNO
				}
			}

			//persiste validadores
			mutex.Lock()
			setValidadores := validadores
			mutex.Unlock()

			//para cada token do validador na stake, insere a identificacao deste validador na loteria
			k, ok := setValidadores[bloco.Validador]
			if ok {
				for i := 0; i < k; i++ {
					loteria = append(loteria, bloco.Validador)
				}
			}
		}

		//escolhe um vencedor aleatório
		if len(loteria) > 0 {
			source := rand.NewSource(time.Now().Unix())
			numero := rand.New(source)
			vencedor := loteria[numero.Intn(len(loteria))]

			for _, bloco := range temp {
				if bloco.Validador == vencedor {
					mutex.Lock()
					Blockchain = append(Blockchain, bloco)
					delete(validadores, bloco.Validador)
					loteria = []string{}
					spew.Dump(Blockchain)
					mutex.Unlock()
					break
				}
			}
		}

	}
	mutex.Lock()
	tempBlocos = []Bloco{}
	mutex.Unlock()
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
		// cart, ok := criaCarteira(false)

		// if ok {
		// 	stk, _ := cart.geraStake(true)
		// 	cart.Stakes = append(cart.Stakes, stk)
		// 	carteiras = append(carteiras, cart)
		// 	spew.Dump(carteiras)
		// }

		for indiceCart, cart := range carteiras {
			for indiceStk, stk := range cart.Stakes {
				spew.Dump(stk)
				if !stk.Sent {
					// stkJSON := []byte(`{"id":"` + stk.ID.String() +
					// 	`", "idcarteira":"` + stk.IDCarteira.String() +
					// 	`", "dados":` + strconv.Itoa(stk.Dados) +
					// 	`, "tokens":` + strconv.Itoa(stk.Tokens) +
					// 	`}`)
					// req, err := http.NewRequest("POST", "http://localhost:9000", bytes.NewBuffer(stkJSON))
					// if err != nil {
					// 	log.Fatal(err)
					// }
					// req.Close = true
					// req.Header.Set("Content-Type", "application/json")
					// client := &http.Client{Timeout: 0 * time.Second}
					// _, err = enviaStake(client, req)
					// if err != nil {
					// 	log.Fatal(err)
					// }
					stakes = append(stakes, stk)
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
		if len(stakes) > 10 {
			for _, cart := range carteiras {

			}
		}
		slp := rand.Intn(20)
		spew.Dump(slp)
		time.Sleep(time.Duration(slp) * time.Second)
	}
}
