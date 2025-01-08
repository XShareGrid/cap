package mysql

import (
	"log"
	"testing"
)

type m1 struct {
}

func (m *m1) Name() string {
	return "m1"
}

func (m *m1) InitMapper(ss *Session) error {
	log.Println("InitMapper:" + m.Name())
	return nil
}

func (m *m1) Dependencies() []string {
	return []string{"m2"}
}

type m2 struct {
}

func (m *m2) Name() string {
	return "m2"
}
func (m *m2) InitMapper(ss *Session) error {
	log.Println("InitMapper:" + m.Name())
	return nil
}

func (m *m2) Dependencies() []string {
	return []string{"m3"}
}

type m3 struct {
}

func (m *m3) Name() string {
	return "m3"
}

func (m *m3) InitMapper(ss *Session) error {
	log.Println("InitMapper:" + m.Name())
	return nil
}

func (m *m3) Dependencies() []string {
	return []string{"m4"}
}

type m4 struct {
}

func (m *m4) Name() string {
	return "m4"
}

func (m *m4) InitMapper(ss *Session) error {
	log.Println("InitMapper:" + m.Name())
	return nil
}

func (m *m4) Dependencies() []string {
	return nil
}

func Test_mapperManager_register(t *testing.T) {
	m := mapperMgr{}
	m.register(&m1{}, &m3{}, &m2{}, &m4{})
	m.initMappers(nil)
}
