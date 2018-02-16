package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/andyleap/casfs/casdist"
	casfs "github.com/andyleap/casfs/fs"
	casfsproto "github.com/andyleap/casfs/fs/proto"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/coreos/etcd/clientv3"
)

var (
	casHost     = flag.String("cas", "127.0.0.1", "CAS Host(s)")
	etcdCluster = flag.String("etcd", "127.0.0.1", "etcd Host")
	fsVol       = flag.String("volume", "", "Volume to attach to")
	debug       = flag.Bool("debug", false, "turn on debug logging")
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("expected 1 arg")
	}
	mountpoint := flag.Arg(0)

	if *debug {
		fuse.Debug = func(msg interface{}) {
			log.Println(msg)
		}
	}

	c, err := fuse.Mount(
		mountpoint,
		fuse.NoAppleDouble(),
		fuse.NoAppleXattr(),
		fuse.FSName("casfs"),
		fuse.Subtype("casfs"),
	)

	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	names, _ := net.LookupAddr(*casHost)
	for i, name := range names {
		names[i] = name + ":3124"
	}
	log.Println(names)
	cas := casdist.New(names, 1)

	names, _ = net.LookupAddr(*etcdCluster)
	for i, name := range names {
		names[i] = name + ":2379"
	}
	client, _ := clientv3.New(clientv3.Config{Endpoints: names})

	cfs := New(cas, client, *fsVol)
	err = fs.Serve(c, cfs)
	if err != nil {
		log.Fatal(err)
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}

	log.Println(hex.EncodeToString(cfs.cfs.Hash()))
}

type FS struct {
	cfs    *casfs.FS
	client *clientv3.Client
}

func New(cas casfs.CAS, client *clientv3.Client, volume string) *FS {
	resp, _ := client.Get(context.Background(), fmt.Sprintf("/casfs/vol/%s/hash", volume))
	hash := []byte{}
	if resp.Count > 0 {
		hash = resp.Kvs[0].Value
	}
	cfs := casfs.New(cas, hash)
	cfs.OnChange = func(hash []byte) {
		client.Put(context.Background(), fmt.Sprintf("/casfs/vol/%s/hash", volume), string(hash))
	}
	return &FS{
		cfs:    cfs,
		client: client,
	}
}

func (fs FS) Root() (fs.Node, error) {
	return (*Dir)(fs.cfs.Root()), nil
}

type Dir casfs.Dir

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("ATTR")
	attr := (*casfs.Dir)(d).Attr()
	a.Mode = os.ModeDir | os.FileMode(attr.Mode)
	a.Uid = attr.UID
	a.Gid = attr.GID
	return nil
}

func (d *Dir) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	log.Println("SETATTR")
	attr := (*casfs.Dir)(d).Attr()
	if req.Valid.Mode() {
		attr.Mode = uint32(req.Mode)
	}
	if req.Valid.Uid() {
		attr.UID = req.Uid
	}
	if req.Valid.Gid() {
		attr.GID = req.Gid
	}
	(*casfs.Dir)(d).SetAttr(attr)
	return nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	dir, err := (*casfs.Dir)(d).CreateDir(req.Name)
	if err == nil {
		attr := casfsproto.Attr{
			GID:  req.Gid,
			UID:  req.Uid,
			Mode: uint32(req.Mode),
		}
		dir.SetAttr(attr)
	}
	wdir := (*Dir)(dir)
	return wdir, err
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("CREATE: ", req.Name)
	if req.Mode.IsDir() {
		dir, err := (*casfs.Dir)(d).CreateDir(req.Name)
		if err == nil {
			attr := casfsproto.Attr{
				GID:  req.Gid,
				UID:  req.Uid,
				Mode: uint32(req.Mode),
			}
			dir.SetAttr(attr)
		}
		if err == casfs.ErrAlreadyExists {
			if req.Flags&fuse.OpenExclusive == 0 {
				dir, err = (*casfs.Dir)(d).OpenDir(req.Name)
			}
		}
		wdir := (*Dir)(dir)
		return wdir, wdir, err
	}
	file, err := (*casfs.Dir)(d).CreateFile(req.Name)
	if err == nil {
		file.SetAttr(casfsproto.Attr{
			Mode: uint32(req.Mode),
			UID:  req.Uid,
			GID:  req.Gid,
		})
	}
	if err == casfs.ErrAlreadyExists {
		if req.Flags&fuse.OpenExclusive == 0 {
			file, err = (*casfs.Dir)(d).OpenFile(req.Name)
		}
	}
	wfile := (*File)(file)
	return wfile, wfile, err
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	return (*casfs.Dir)(d).Delete(req.Name)
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	dir, err := (*casfs.Dir)(d).OpenDir(name)
	if err == nil {
		wdir := (*Dir)(dir)
		return wdir, err
	}
	file, err := (*casfs.Dir)(d).OpenFile(name)
	wfile := (*File)(file)
	if err == casfs.ErrNotFound {
		return nil, fuse.ENOENT
	}
	return wfile, err
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	items := (*casfs.Dir)(d).ReadDir()
	ret := make([]fuse.Dirent, len(items))
	for i, item := range items {
		ret[i].Name = item.Name
		ret[i].Type = fuse.DT_File
		if item.Type == casfsproto.ItemDir {
			ret[i].Type = fuse.DT_Dir
		}
	}
	return ret, nil
}

type File casfs.File

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	attr := (*casfs.File)(f).Attr()
	a.Mode = os.FileMode(attr.Mode)
	a.Size = (*casfs.File)(f).Size()
	a.Uid = attr.UID
	a.Gid = attr.GID
	return nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	attr := (*casfs.File)(f).Attr()
	if req.Valid.Mode() {
		attr.Mode = uint32(req.Mode)
	}
	if req.Valid.Uid() {
		attr.UID = req.Uid
	}
	if req.Valid.Gid() {
		attr.GID = req.Gid
	}
	(*casfs.File)(f).SetAttr(attr)
	if req.Valid.Size() {
		(*casfs.File)(f).Truncate(int64(req.Size))
	}
	return nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if cap(resp.Data) < req.Size {
		resp.Data = make([]byte, req.Size)
	} else {
		resp.Data = resp.Data[:req.Size]
	}
	n, err := (*casfs.File)(f).ReadAt(resp.Data, req.Offset)
	resp.Data = resp.Data[:n]
	return err
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) (err error) {
	resp.Size, err = (*casfs.File)(f).WriteAt(req.Data, req.Offset)
	return
}
