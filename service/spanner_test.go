package service

import (
	"fmt"
	"testing"
)

// func TestWriteUsingDML(t *testing.T) {
// 	writeUsingDML("projects/emu/instances/emu/databases/emu")
// }

func TestGetSingerList(t *testing.T) {
	singerList, err := getSingerList("projects/emu/instances/emu/databases/emu")
	if err != nil {
		fmt.Printf("err %s", err)

	}

	for _, singer := range singerList {
		fmt.Printf("singer %s", singer.FirstName)
	}
}

func TestGetSandboxList(t *testing.T) {
	entityList, err := getSandboxList("projects/emu/instances/emu/databases/emu")
	if err != nil {
		fmt.Printf("err %s", err)
	}

	for _, entity := range entityList {
		fmt.Printf("entity %d", entity.IntCl)
	}
}
