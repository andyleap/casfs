package main

import (
	"flag"
	"log"
	"encoding/hex"

	"github.com/andyleap/casfs/cas"
	"github.com/andyleap/casfs/fs"
	fsproto "github.com/andyleap/casfs/fs/proto"
)

var (
	casServer = flag.String("cas", "127.0.0.1:3124", "CAS server to GC")
)

func main() {
	flag.Parse()
	c, err := cas.Dial(*casServer)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Marking")
	c.Mark()
	log.Println("Marked")

	for _, hash := range flag.Args() {
		rawHash, _ := hex.DecodeString(hash)
		cfs := fs.New(c, rawHash)
		WalkDir(cfs.Root())
	}

	log.Println("Sweeping")
	c.Sweep()
	log.Println("Sweeped")
}

func WalkDir(dir *fs.Dir) {
	for _, item := range dir.ReadDir() {
		switch item.Type {
		case fsproto.ItemDir:
			d, err := dir.OpenDir(item.Name)
			if err != nil {
				log.Fatal(err)
			}
			WalkDir(d)
		case fsproto.ItemFile:
			f, err := dir.OpenFile(item.Name)
			if err != nil {
				log.Fatal(err)
			}
			WalkFile(f)
		}
	}
}

func WalkFile(file *fs.File) {
	b := make([]byte, 40960)
	for l1 := int64(0); l1 < int64(file.Size()); l1 += int64(len(b)) {
		file.ReadAt(b, l1)
	}
}
