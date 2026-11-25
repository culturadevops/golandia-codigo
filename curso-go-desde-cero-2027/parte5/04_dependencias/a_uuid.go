package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	id := uuid.NewString()
	fmt.Println("largo del id:", len(id), "· versión:", uuid.MustParse(id).Version())
	fijo := uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://golandia.dev"))
	fmt.Println("id determinista:", fijo)
}
