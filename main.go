package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	// "fmt"

	"math/rand"
	"time"

	"github.com/davecgh/go-spew/spew"
	cmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	wr "github.com/mroth/weightedrand"
)

type Carteira struct {
	ID       uuid.UUID
	Stakes   []Stake
	Currency int
}

type Stake struct {
	ID                uuid.UUID
	IDCarteiraOrigem  uuid.UUID
	IDCarteiraDestino uuid.UUID
	Currency          int
}

type Bloco struct {
	Indice    int
	Timestamp string
	Dados     []Stake
	Hash      string
	HashAnt   string
	Validador string
}

var (
	Blockchain                 []Bloco
	stakes                     []Stake
	validadores                = make(map[string]int)
	anunciador                 = make(chan string)
	mutexValor                 = &sync.Mutex{}
	mutexStks                  = &sync.Mutex{}
	blocosConsolidadosTX       = 0
	blocosConsolidadosCurrency = make(map[string]int)
	valorRetidoEsperandoTx     = make(map[string]int)
	carteiras                  []*Carteira
	mutex                      = &sync.Mutex{}
)

func calculaHash(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	hashFinal := hasher.Sum(nil)
	return hex.EncodeToString(hashFinal)
}

func calculaHashBloco(bloco Bloco) string {
	totalDados := strconv.Itoa(bloco.Indice) + bloco.Timestamp + bloco.HashAnt
	return calculaHash(totalDados)
}

func geraBloco(blocoAnterior Bloco, validador string, stks []Stake) (Bloco, error) {
	var novoBloco Bloco

	t := time.Now()

	novoBloco.Indice = blocoAnterior.Indice + 1
	novoBloco.Timestamp = t.String()
	novoBloco.Dados = stks
	novoBloco.HashAnt = blocoAnterior.Hash
	novoBloco.Hash = calculaHashBloco(novoBloco)
	novoBloco.Validador = validador

	return novoBloco, nil
}

func (cart *Carteira) atualizaCarteira() {
	for _, blc := range Blockchain[blocosConsolidadosCurrency[cart.ID.String()]:] {
		mutexValor.Lock()
		blocosConsolidadosCurrency[cart.ID.String()]++
		cart.Currency -= valorRetidoEsperandoTx[cart.ID.String()]
		valorRetidoEsperandoTx[cart.ID.String()] = 0
		mutexValor.Unlock()
		if blc.Validador == cart.ID.String() {
			tip := 0
			for _, stk := range blc.Dados {
				tip += stk.Currency
			}
			tip = tip / 10
			cart.Currency += tip
		}
		for _, stake := range blc.Dados {
			switch id := cart.ID; id {
			case stake.IDCarteiraDestino:
				cart.Currency += stake.Currency
			}
		}
	}
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
	loteria := make(map[string]int)
	if len(stakes) > 0 {
		for _, stk := range stakes[0:15] {
			if _, ok := loteria[stk.IDCarteiraOrigem.String()]; !ok {
				loteria[stk.IDCarteiraOrigem.String()] = stk.Currency
			} else {
				loteria[stk.IDCarteiraOrigem.String()] += stk.Currency
			}
		}

		//escolhe um vencedor aleatório
		if len(loteria) > 0 {
			rand.Seed(time.Now().UTC().UnixNano())
			var choices []wr.Choice
			for k, v := range loteria {
				choices = append(choices, wr.NewChoice(k, uint(v)))
			}
			chooser, _ := wr.NewChooser(choices...)
			novoBloco, _ := geraBloco(Blockchain[len(Blockchain)-1], chooser.Pick().(string), stakes[0:15])
			_ = insertBloco(novoBloco)
			if len(Blockchain)%100 == 0 {
				salvaEstado()
			}
		}
	}
}

func criaCarteira(inicial bool) (cart Carteira, ok bool) {
	source := rand.NewSource(time.Now().Unix())
	numero := rand.New(source)
	if inicial {
		cart := Carteira{
			ID:       uuid.New(),
			Currency: 100,
		}
		return cart, true
	}
	if numero.Intn(100) > 79 {
		cart := Carteira{
			ID:       uuid.New(),
			Currency: 100,
		}
		return cart, true
	}
	return Carteira{}, false
}

func (cart *Carteira) geraStake(carteiras []*Carteira) Stake {
	rand.Seed(time.Now().UnixNano())
	t := Stake{}
	mutexValor.Lock()
	c := carteiras[rand.Intn(len(carteiras))]
	if (c.ID != cart.ID) && ((cart.Currency - valorRetidoEsperandoTx[cart.ID.String()]) > 0) {
		valor := rand.Intn(cart.Currency - valorRetidoEsperandoTx[cart.ID.String()])
		if valor > 0 {
			valorRetidoEsperandoTx[cart.ID.String()] += valor
			t.ID = uuid.New()
			t.IDCarteiraDestino = c.ID
			t.IDCarteiraOrigem = cart.ID
			t.Currency = valor
			mutexValor.Unlock()
			return t
		}
	}
	mutexValor.Unlock()
	return t
}

func insertBloco(novoBloco Bloco) bool {
	if (!cmp.Equal(novoBloco, Bloco{})) && blocoValido(novoBloco, Blockchain[len(Blockchain)-1]) {
		mutex.Lock()
		Blockchain = append(Blockchain, novoBloco)
		limpaStakes()
		mutex.Unlock()
		return true
	}
	return false
}

func limpaStakes() {
	mapTransacoes := make(map[uuid.UUID]Stake)
	for _, t := range stakes {
		mapTransacoes[t.ID] = t
	}
	blocos := 0
	for _, bloco := range Blockchain[blocosConsolidadosTX:] {
		blocos++
		for _, trans := range bloco.Dados {
			for _, tfila := range stakes {
				if trans.ID == tfila.ID {
					delete(mapTransacoes, tfila.ID)
				}
			}
		}
	}
	blocosConsolidadosTX += blocos
	txs := make([]Stake, 0, len(mapTransacoes))
	for _, tx := range mapTransacoes {
		txs = append(txs, tx)
	}
	mutexStks.Lock()
	stakes = txs
	mutexStks.Unlock()
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Servlet ouvindo na porta ", httpAddr)
	server := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

//funcao que cria o roteador
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/stakes", handleGetStakes).Methods("GET")
	muxRouter.HandleFunc("/carteiras", handleGetCarteiras).Methods("GET")
	return muxRouter
}

func handleGetBlockchain(writer http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	mutex.Unlock()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, string(bytes))
}

func handleGetStakes(writer http.ResponseWriter, req *http.Request) {
	mutexStks.Lock()
	bytes, err := json.MarshalIndent(stakes, "", "  ")
	mutexStks.Unlock()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, string(bytes))
}

func handleGetCarteiras(writer http.ResponseWriter, req *http.Request) {
	bytes, err := json.MarshalIndent(carteiras, "", "  ")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, string(bytes))
}

func startCarteiras() {
	for {
		for _, cart := range carteiras {
			cart.atualizaCarteira()
			time.Sleep(600 * time.Millisecond)
			stk := cart.geraStake(carteiras)
			if (stk != Stake{}) {
				mutexStks.Lock()
				stakes = append(stakes, stk)
				mutexStks.Unlock()
			}
		}
	}
}

func salvaEstado() {
	spew.Dump("Salvando... " + time.Now().Format("15:04:05"))
	file1 := "./carteiras.json"
	file4 := "./blockchain.json"

	bytes1, _ := json.MarshalIndent(carteiras, "", "  ")
	bytes4, _ := json.MarshalIndent(Blockchain, "", "  ")

	_ = ioutil.WriteFile(file1, bytes1, 0644)
	_ = ioutil.WriteFile(file4, bytes4, 0644)
	spew.Dump("Salvo! " + time.Now().Format("15:04:05"))
}

func main() {
	for i := 0; i < 150; i++ {
		cart, _ := criaCarteira(true)
		carteiras = append(carteiras, &cart)
	}
	spew.Dump(carteiras)
	go func() {
		for {
			if len(stakes) > 15 {
				escolheValidador()
			}
			time.Sleep(15 * time.Second)
		}
	}()
	t := time.Now()
	blocoGenese := Bloco{}
	hasher := sha256.New()
	blocoGenese = Bloco{0, t.String(), []Stake{}, "", "", ""}
	totalDados := strconv.Itoa(blocoGenese.Indice) + blocoGenese.Timestamp + blocoGenese.HashAnt
	hasher.Write([]byte(totalDados))
	hashFinal := hasher.Sum(nil)
	blocoGenese.Hash = hex.EncodeToString(hashFinal)

	spew.Dump(blocoGenese)

	mutex.Lock()
	Blockchain = append(Blockchain, blocoGenese)
	mutex.Unlock()
	_ = godotenv.Load()
	go startCarteiras()
	log.Fatal(run())
}
