package main

import (
	"context"
	cr "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
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

const (
	url = "http://localhost:8080"
)

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
	blockchain                 []Bloco
	cartMaliciosas             = make(map[string]*Carteira)
	malicious                  bool
	verbose                    int
	totalcart                  int
	dificuldade                int
	alvoBlocos                 int
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
	Malicious   bool
}

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
	mutexBC.Lock()
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
	mutexBC.Unlock()
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

func geraBloco(ctx context.Context, blocoAntigo Bloco, transacoes []Transacao, dificuldade int, idCart string, malicious bool) Bloco {
	var novoBloco Bloco

	t := time.Now()

	novoBloco.Indice = blocoAntigo.Indice + 1
	novoBloco.Timestamp = t.String()
	novoBloco.Transacoes = transacoes
	novoBloco.HashAnt = blocoAntigo.Hash
	novoBloco.Dificuldade = dificuldade
	novoBloco.Minerador = idCart
	novoBloco.Malicious = malicious

	for {
		select {
		case <-ctx.Done():
			if verbose >= 3 {
				spew.Dump("ctx.Cancel()")
			}
			return Bloco{}
		default:
			hex := fmt.Sprintf("%x", rand.Intn(int(^uint32(0))))
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
	cart.atualizaCarteira()
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

func getMaliciousTXs() []Transacao {
	var retorno []Transacao
	for _, tx := range transactions {
		if _, ok := cartMaliciosas[tx.IDCarteiraOrigem.String()]; ok {
			retorno = append(retorno, tx)
		}
	}
	return retorno
}

func (cart *Carteira) start(ctx context.Context, cancel *context.CancelFunc) {
	// cart.atualizaCarteira()
	// sleep := rand.Intn(500)
	// time.Sleep(time.Millisecond * time.Duration(sleep))
	if true || rand.Intn(100) >= 70 {
		transacao := cart.criaTransacao(carteiras)
		if (!cmp.Equal(transacao, Transacao{})) {
			if verbose >= 2 {
				spew.Dump(cart.ID.String() + " gerou transação.")
			}
			mutexTrans.Lock()
			transactions = append(transactions, transacao)
			mutexTrans.Unlock()
		}
	}

	if len(transactions) > 15 {
		if true || rand.Intn(100) >= 50 {
			if _, ok := cartMaliciosas[cart.ID.String()]; !ok || !malicious {
				mutexBC.Lock()
				novoBloco := geraBloco(ctx, blockchain[len(blockchain)-1], transactions, dificuldade, cart.ID.String(), false)
				if insertBloco(novoBloco) {
					mutexBC.Unlock()
					if verbose >= 1 {
						spew.Dump("Bloco inserido com sucesso. " + time.Now().Format("15:04:05"))
					}
					c := *cancel
					c()
				}
				mutexBC.Unlock()
			} else if ok {
				if txs := getMaliciousTXs(); !cmp.Equal(txs, []Transacao{}) {
					mutexBC.Lock()
					novoBloco := geraBloco(ctx, blockchain[len(blockchain)-1], txs, dificuldade, cart.ID.String(), true)
					if insertBloco(novoBloco) {
						mutexBC.Unlock()

						if verbose >= 1 {
							spew.Dump("Bloco malicioso inserido com sucesso. " + time.Now().Format("15:04:05"))
						}
						c := *cancel
						c()
					}
					mutexBC.Unlock()
				}
			}
		}
	}
	working.Store(cart.ID.String(), false)
	return
}

func insertBloco(novoBloco Bloco) bool {
	if (!cmp.Equal(novoBloco, Bloco{})) && blocoValido(novoBloco, blockchain[len(blockchain)-1]) {
		// mutexBC.Lock()
		blockchain = append(blockchain, novoBloco)
		if len(blockchain)%100 == 0 {
			salvaEstado()
		}
		mutexTrans.Lock()
		limpaTransacoes()
		mutexTrans.Unlock()
		// mutexBC.Unlock()
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
		if len(blockchain) > alvoBlocos {
			os.Exit(1)
		}
		select {
		case <-ctx.Done():
			ctx, cancel = context.WithCancel(context.Background())
		default:
			for _, cart := range carteiras {
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
	muxRouter.HandleFunc("/malicious", handleGetCarteirasMaliciosas).Methods("GET")
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

func handleGetCarteirasMaliciosas(writer http.ResponseWriter, req *http.Request) {
	bytes, err := json.MarshalIndent(cartMaliciosas, "", "  ")
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

func salvaEstado() {
	spew.Dump("Salvando... " + time.Now().Format("15:04:05"))
	file1 := "./carteiras.json"
	file2 := "./malicious.json"
	file3 := "./transactions.json"
	file4 := "./blockchain.json"

	bytes1, _ := json.MarshalIndent(carteiras, "", "  ")
	bytes2, _ := json.MarshalIndent(cartMaliciosas, "", "  ")
	bytes3, _ := json.MarshalIndent(transactions, "", "  ")
	bytes4, _ := json.MarshalIndent(blockchain, "", "  ")

	_ = ioutil.WriteFile(file1, bytes1, 0644)
	_ = ioutil.WriteFile(file2, bytes2, 0644)
	_ = ioutil.WriteFile(file3, bytes3, 0644)
	_ = ioutil.WriteFile(file4, bytes4, 0644)
	spew.Dump("Salvo! " + time.Now().Format("15:04:05"))
}

func init() {
	var b [8]byte
	_, err := cr.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func main() {
	//Carrega variáveis de ambiente
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	totalcart, _ = strconv.Atoi(os.Args[1])
	dificuldade, _ = strconv.Atoi(os.Args[2])
	malicious, _ = strconv.ParseBool(os.Args[3])
	verbose, _ = strconv.Atoi(os.Args[4])
	alvoBlocos, _ = strconv.Atoi(os.Args[5])
	//Cria 100 carteiras, cada uma com 100 moedas em seu estado inicial.
	for numcart := 0; numcart < totalcart; numcart++ {
		carteiras = append(carteiras, &Carteira{
			ID:       uuid.New(),
			Currency: 100,
		})
	}
	if malicious {
		for numcart := 0; numcart <= int(math.Ceil((float64(totalcart) * .51))); numcart++ {
			cart := carteiras[numcart]
			cartMaliciosas[cart.ID.String()] = cart
		}
		spew.Dump(cartMaliciosas)
	}
	//Criação do bloco gênese
	t := time.Now()
	blocoGenese := Bloco{}
	hasher := sha256.New()
	blocoGenese = Bloco{0, t.String(), []Transacao{}, "", "", 0, "", "", false}
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
