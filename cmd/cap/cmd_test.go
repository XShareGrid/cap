package main

import "testing"

func Test_table2Go(t *testing.T) {
	table2Go("root:125801@tcp(localhost:3306)/cap?charset=utf8", "test.go", "mysql", "", true)
}
