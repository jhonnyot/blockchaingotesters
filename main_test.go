package main

import (
	"testing"
)

func BenchmarkMain(b *testing.B) {
	// its, numerros := 0, 0
	// var erros []error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// its = its + 1
			main()
			// 	if err != nil {
			// 		numerros = numerros + 1
			// 		erros = append(erros, err)
			// 	}
		}
	})
}
