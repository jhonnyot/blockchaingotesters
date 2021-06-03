package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	cmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const url = "http://localhost:8080"

var (
	carteiras                  []*Carteira
	working                    = sync.Map{}
	mutexBC                    = &sync.Mutex{}
	mutexTrans                 = &sync.Mutex{}
	mutexValor                 = &sync.Mutex{}
	transactions               []Transacao
	blocosConsolidadosTX       = 0
	blocosConsolidadosCurrency = make(map[string]int)
	valorRetidoEsperandoTx     = make(map[string]int)
)

type Bloco struct {
	Indice      int
	Timestamp   string
	Transacoes  []Transacao
	Hash        string
	HashAnt     string
	Dificuldade int
	Nonce       string
	Minerador   string
}

var blockchain []Bloco

type Transacao struct {
	ID                uuid.UUID
	IDCarteiraOrigem  uuid.UUID
	IDCarteiraDestino uuid.UUID
	Currency          int
}

type Carteira struct {
	ID       uuid.UUID
	Currency int
}

func (cart *Carteira) atualizaCarteira() {
	for _, blc := range blockchain[blocosConsolidadosCurrency[cart.ID.String()]:] {
		mutexValor.Lock()
		blocosConsolidadosCurrency[cart.ID.String()]++
		cart.Currency -= valorRetidoEsperandoTx[cart.ID.String()]
		valorRetidoEsperandoTx[cart.ID.String()] = 0
		mutexValor.Unlock()
		if blc.Minerador == cart.ID.String() {
			cart.Currency += 100
		}
		for _, tsc := range blc.Transacoes {
			switch id := cart.ID; id {
			case tsc.IDCarteiraDestino:
				cart.Currency += tsc.Currency
			}
		}
	}
}

func calculaHash(bloco Bloco) string {
	totalDados := strconv.Itoa(bloco.Indice) + bloco.Timestamp + bloco.HashAnt + bloco.Nonce
	hasher := sha256.New()
	hasher.Write([]byte(totalDados))
	hashFinal := hasher.Sum(nil)
	return hex.EncodeToString(hashFinal)
}

func validaHash(hash string, dificuldade int) bool {
	prefixo := strings.Repeat("0", dificuldade)
	return strings.HasPrefix(hash, prefixo)
}

func geraBloco(ctx context.Context, blocoAntigo Bloco, transacoes []Transacao, dificuldade int, idCart string) Bloco {
	var novoBloco Bloco

	t := time.Now()

	novoBloco.Indice = blocoAntigo.Indice + 1
	novoBloco.Timestamp = t.String()
	novoBloco.Transacoes = transacoes
	novoBloco.HashAnt = blocoAntigo.Hash
	novoBloco.Dificuldade = dificuldade
	novoBloco.Minerador = idCart

	for {
		select {
		case <-ctx.Done():
			spew.Dump("ctx.Cancel()")
			return Bloco{}
		default:
			src := rand.NewSource(time.Now().UnixNano())
			r := rand.New(src)
			hex := fmt.Sprintf("%x", r.Intn(int(^uint32(0))))
			novoBloco.Nonce = hex
			if !validaHash(calculaHash(novoBloco), novoBloco.Dificuldade) {
				continue
			} else {
				fmt.Println(idCart, "gerou bloco válido.")
				novoBloco.Hash = calculaHash(novoBloco)
				break
			}
		}
		return novoBloco
	}
}

func (cart *Carteira) criaTransacao(carteiras []*Carteira) Transacao {
	rand.Seed(time.Now().Unix())
	t := Transacao{}
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

func (cart *Carteira) start(ctx context.Context, cancel *context.CancelFunc) {
	rand.Seed(time.Now().Unix())
	if rand.Intn(100) >= 70 {
		transacao := cart.criaTransacao(carteiras)
		if (!cmp.Equal(transacao, Transacao{})) {
			spew.Dump(cart.ID.String() + " gerou transação.")
			mutexTrans.Lock()
			transactions = append(transactions, transacao)
			mutexTrans.Unlock()
		}
	}

	if len(transactions) > 99 {
		novoBloco := geraBloco(ctx, blockchain[len(blockchain)-1], transactions, 4, cart.ID.String())
		if insertBloco(novoBloco) {
			spew.Dump("Bloco inserido com sucesso.")
			c := *cancel
			c()
		}
	}
	working.Store(cart.ID.String(), false)
	return
}

func insertBloco(novoBloco Bloco) bool {
	if (!cmp.Equal(novoBloco, Bloco{})) && blocoValido(novoBloco, blockchain[len(blockchain)-1]) {
		mutexBC.Lock()
		blockchain = append(blockchain, novoBloco)
		limpaTransacoes()
		mutexBC.Unlock()
		return true
	}
	return false
}

func blocoValido(novoBloco, blocoAnterior Bloco) bool {
	if blocoAnterior.Indice+1 != novoBloco.Indice {
		return false
	}
	if blocoAnterior.Hash != novoBloco.HashAnt {
		return false
	}

	return true
}

func startCarteiras() {
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case <-ctx.Done():
			ctx, cancel = context.WithCancel(context.Background())
		default:
			for _, cart := range carteiras {
				cart.atualizaCarteira()
				if _, found := working.Load(cart.ID.String()); !found {
					working.Store(cart.ID.String(), false)
				}
				if v, _ := working.Load(cart.ID.String()); !v.(interface{}).(bool) {
					working.Store(cart.ID.String(), true)
					go cart.start(ctx, &cancel)
				}
			}
		}
	}
}

func run() error {
	mux := makeMuxRouter()
	//Recupera a porta a ser usada pelo servidor nas variáveis de ambiente
	httpAddr := os.Getenv("ADDR")
	log.Println("Servlet ouvindo na porta ", httpAddr)
	//Define o servidor HTTP
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
	muxRouter.HandleFunc("/transactions", handleGetTransactions).Methods("GET")
	muxRouter.HandleFunc("/carteiras", handleGetCarteiras).Methods("GET")
	return muxRouter
}

func handleGetBlockchain(writer http.ResponseWriter, req *http.Request) {
	mutexBC.Lock()
	bytes, err := json.MarshalIndent(blockchain, "", "  ")
	mutexBC.Unlock()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	io.WriteString(writer, string(bytes))
}

func handleGetTransactions(writer http.ResponseWriter, req *http.Request) {
	mutexTrans.Lock()
	bytes, err := json.MarshalIndent(transactions, "", "  ")
	mutexTrans.Unlock()
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

func limpaTransacoes() {
	//Cria e popula um mapa com todas as transações da pool
	mapTransacoes := make(map[uuid.UUID]Transacao)
	for _, t := range transactions {
		mapTransacoes[t.ID] = t
	}
	blocos := 0
	//Percorre os blocos a partir do último bloco consolidado
	for _, bloco := range blockchain[blocosConsolidadosTX:] {
		blocos++
		//Percorre as transações destes blocos
		for _, trans := range bloco.Transacoes {
			//Retira do mapa as transações que aparecem nestes blocos
			for _, tfila := range transactions {
				if trans.ID == tfila.ID {
					delete(mapTransacoes, tfila.ID)
				}
			}
		}
	}
	//Consolida os blocos percorridos
	blocosConsolidadosTX += blocos
	//Cria uma slice auxiliar de transações
	txs := make([]Transacao, 0, len(mapTransacoes))
	//Adiciona nesta slice as transações restantes do mapa
	for _, tx := range mapTransacoes {
		txs = append(txs, tx)
	}
	//Transforma a slice auxiliar na nova pool de transações
	transactions = txs
}

func main() {
	//Carrega variáveis de ambiente
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	//Cria 100 carteiras, cada uma com 100 moedas em seu estado inicial.
	for numcart := 0; numcart < 100; numcart++ {
		carteiras = append(carteiras, &Carteira{
			ID:       uuid.New(),
			Currency: 100,
		})
	}
	//Criação do bloco gênese
	t := time.Now()
	blocoGenese := Bloco{}
	hasher := sha256.New()
	blocoGenese = Bloco{0, t.String(), []Transacao{}, "", "", 0, "", ""}
	totalDados := strconv.Itoa(blocoGenese.Indice) + blocoGenese.Timestamp + blocoGenese.HashAnt + blocoGenese.Nonce
	hasher.Write([]byte(totalDados))
	hashFinal := hasher.Sum(nil)
	blocoGenese.Hash = hex.EncodeToString(hashFinal)
	//Insere o bloco gênese na blockchain
	mutexBC.Lock()
	blockchain = append(blockchain, blocoGenese)
	mutexBC.Unlock()
	//goroutine que começa o trabalho das carteiras
	go startCarteiras()
	//levanta o servidor http
	log.Fatal(run())
}
