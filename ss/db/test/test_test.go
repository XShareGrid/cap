package t

import (
	"testing"
)

func Test_testSQLTxTimeout(t *testing.T) {
	testSQLTxTimeout()
}

func Test_testSQLTxTimeoutWithClean(t *testing.T) {
	testSQLTxTimeoutWithClean()
}

func Test_testSQLTxPanic(t *testing.T) {
	testSQLTxPanic()
}
