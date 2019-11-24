package main

import (
	"log"
	"strconv"
	"testing"
)

// func BenchmarkPostReq(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		err := postRequest()
// 		if err != nil {
// 			b.Errorf("Erro nos requests.")
// 		}
// 	}
// }

func BenchmarkParaleloDificuldade6(b *testing.B) {
	its, numerros := 0, 0
	var erros []error
	var mapaErros map[string]int
	mapaErros = make(map[string]int)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			its = its + 1
			err := postRequest(5)
			if err != nil {
				numerros = numerros + 1
				erros = append(erros, err)
				msg := err.Error()
				mapaErros[msg] = mapaErros[msg] + 1
			}
		}
	})
	if its > 1 {
		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
	}
	// for _, msg := range erros {
	// 	if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
	// 		defer log.Println(msg)
	// 	}
	for key, value := range mapaErros {
		defer log.Printf("Mensagem: %v; Número: %v", key, strconv.Itoa(value))
	}
}
