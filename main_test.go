package main

import (
	"log"
	"strconv"
	"strings"
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

func BenchmarkParaleloDificuldade3(b *testing.B) {
	its, numerros := 0, 0
	var erros []error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			its = its + 1
			err := postRequest(3)
			if err != nil {
				numerros = numerros + 1
				erros = append(erros, err)
			}
		}
	})
	if its > 1 {
		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
	}
	for _, msg := range erros {
		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
			defer log.Println(msg)
		}
	}
}

func BenchmarkParaleloDificuldade4(b *testing.B) {
	its, numerros := 0, 0
	var erros []error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			its = its + 1
			err := postRequest(4)
			if err != nil {
				numerros = numerros + 1
				erros = append(erros, err)
			}
		}
	})
	if its > 1 {
		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
	}
	for _, msg := range erros {
		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
			defer log.Println(msg)
		}
	}
}

func BenchmarkParaleloDificuldade5(b *testing.B) {
	its, numerros := 0, 0
	var erros []error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			its = its + 1
			err := postRequest(5)
			if err != nil {
				numerros = numerros + 1
				erros = append(erros, err)
			}
		}
	})
	if its > 1 {
		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
	}
	for _, msg := range erros {
		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
			defer log.Println(msg)
		}
	}
}

func BenchmarkParaleloDificuldade6(b *testing.B) {
	its, numerros := 0, 0
	var erros []error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			its = its + 1
			err := postRequest(6)
			if err != nil {
				numerros = numerros + 1
				erros = append(erros, err)
			}
		}
	})
	if its > 1 {
		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
	}
	for _, msg := range erros {
		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
			defer log.Println(msg)
		}
	}
}

func BenchmarkParaleloDificuldade7(b *testing.B) {
	its, numerros := 0, 0
	var erros []error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			its = its + 1
			err := postRequest(7)
			if err != nil {
				numerros = numerros + 1
				erros = append(erros, err)
			}
		}
	})
	if its > 1 {
		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
	}
	for _, msg := range erros {
		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
			defer log.Println(msg)
		}
	}
}

// func BenchmarkParaleloDificuldade8(b *testing.B) {
// 	its, numerros := 0, 0
// 	var erros []error
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			its = its + 1
// 			err := postRequest()
// 			if err != nil {
// 				numerros = numerros + 1
// 				erros = append(erros, err)
// 			}
// 		}
// 	})
// 	if its > 1 {
// 		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
// 	}
// 	for _, msg := range erros {
// 		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
// 			defer log.Println(msg)
// 		}
// 	}
// }

// func BenchmarkParaleloDificuldade9(b *testing.B) {
// 	its, numerros := 0, 0
// 	var erros []error
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			its = its + 1
// 			err := postRequest()
// 			if err != nil {
// 				numerros = numerros + 1
// 				erros = append(erros, err)
// 			}
// 		}
// 	})
// 	if its > 1 {
// 		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
// 	}
// 	for _, msg := range erros {
// 		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
// 			defer log.Println(msg)
// 		}
// 	}
// }

// func BenchmarkParaleloDificuldade10(b *testing.B) {
// 	its, numerros := 0, 0
// 	var erros []error
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			its = its + 1
// 			err := postRequest()
// 			if err != nil {
// 				numerros = numerros + 1
// 				erros = append(erros, err)
// 			}
// 		}
// 	})
// 	if its > 1 {
// 		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
// 	}
// 	for _, msg := range erros {
// 		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
// 			defer log.Println(msg)
// 		}
// 	}
// }

// func BenchmarkParaleloDificuldade11(b *testing.B) {
// 	its, numerros := 0, 0
// 	var erros []error
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			its = its + 1
// 			err := postRequest()
// 			if err != nil {
// 				numerros = numerros + 1
// 				erros = append(erros, err)
// 			}
// 		}
// 	})
// 	if its > 1 {
// 		defer log.Printf("Iterações: %v; Erros: %v", strconv.Itoa(its), strconv.Itoa(numerros)+"\n\r")
// 	}
// 	for _, msg := range erros {
// 		if strings.Compare(msg.Error(), "Post http://localhost:8080: EOF") != 0 {
// 			defer log.Println(msg)
// 		}
// 	}
// }
