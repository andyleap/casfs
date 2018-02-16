package fs

import (
	"errors"
	"io"
	"sync"

	"github.com/andyleap/casfs/fs/proto"
)

var (
	ErrNotFound      = errors.New("Not Found")
	ErrAlreadyExists = errors.New("Already Exists")
)

type FS struct {
	mu       sync.Mutex
	cas      CAS
	root     []byte
	dir      *Dir
	OnChange func(hash []byte)
}

func New(cas CAS, root []byte) *FS {
	fs := &FS{
		cas:  cas,
		root: root,
	}
	fs.dir, _ = fs.loadDir(root, fs)
	return fs
}

func (fs *FS) update() {
	fs.root = fs.dir.hash
	if fs.OnChange != nil {
		fs.OnChange(fs.root)
	}
}

func (fs *FS) Root() *Dir {
	return fs.dir
}

func (fs *FS) Hash() []byte {
	return fs.dir.hash
}

type updater interface {
	update()
}

type node interface {
	nodeHash() []byte
}

type Dir struct {
	fs     *FS
	hash   []byte
	parent updater
	cache  proto.Directory
	open   map[string]node
}

func (dir *Dir) nodeHash() []byte {
	return dir.hash
}

func (dir *Dir) update() {
	for k, v := range dir.cache.Items {
		if node, ok := dir.open[v.Name]; ok {
			dir.cache.Items[k].Hash = node.nodeHash()
		}
	}
	data, _ := dir.cache.Marshal(nil)
	dir.hash, _ = dir.fs.cas.Write(data)
	dir.parent.update()
}

func (fs *FS) loadDir(hash []byte, parent updater) (*Dir, error) {
	if len(hash) == 0 {
		return &Dir{
			fs:     fs,
			parent: parent,
			open:   map[string]node{},
		}, nil
	}
	data, err := fs.cas.Read(hash)
	if err != nil {
		return nil, err
	}
	dirInfo := proto.Directory{}
	_, err = dirInfo.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return &Dir{
		fs:     fs,
		hash:   hash,
		parent: parent,
		cache:  dirInfo,
		open:   map[string]node{},
	}, nil
}

func (fs *FS) loadFile(hash []byte, parent updater) (*File, error) {
	data, err := fs.cas.Read(hash)
	if err != nil {
		return nil, err
	}
	fileInfo := proto.File{}
	_, err = fileInfo.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return &File{
		fs:     fs,
		hash:   hash,
		parent: parent,
		cache:  fileInfo,
	}, nil

}

func (dir *Dir) OpenDir(name string) (*Dir, error) {
	dir.fs.mu.Lock()
	defer dir.fs.mu.Unlock()
	if node, ok := dir.open[name]; ok {
		if d, ok := node.(*Dir); ok {
			return d, nil
		}
		return nil, ErrNotFound
	}
	for _, i := range dir.cache.Items {
		if i.Type == proto.ItemDir {
			if i.Name == name {
				d, err := dir.fs.loadDir(i.Hash, dir)
				if err != nil {
					return nil, err
				}
				dir.open[name] = d
				return d, nil
			}
		}
	}
	return nil, ErrNotFound
}

func (dir *Dir) CreateDir(name string) (*Dir, error) {
	dir.fs.mu.Lock()
	defer dir.fs.mu.Unlock()
	for _, i := range dir.cache.Items {
		if i.Name == name {
			return nil, ErrAlreadyExists
		}
	}
	d := &Dir{
		fs:     dir.fs,
		parent: dir,
		open:   map[string]node{},
	}
	dir.cache.Items = append(dir.cache.Items, proto.Item{Name: name, Type: proto.ItemDir})
	dir.open[name] = d
	dir.update()
	return d, nil
}

func (dir *Dir) OpenFile(name string) (*File, error) {
	dir.fs.mu.Lock()
	defer dir.fs.mu.Unlock()
	if node, ok := dir.open[name]; ok {
		if d, ok := node.(*File); ok {
			return d, nil
		}
		return nil, ErrNotFound
	}
	for _, i := range dir.cache.Items {
		if i.Type == proto.ItemFile {
			if i.Name == name {
				f, err := dir.fs.loadFile(i.Hash, dir)
				if err != nil {
					return nil, err
				}
				dir.open[name] = f
				return f, nil
			}
		}
	}
	return nil, ErrNotFound
}

func (dir *Dir) CreateFile(name string) (*File, error) {
	dir.fs.mu.Lock()
	defer dir.fs.mu.Unlock()
	for _, i := range dir.cache.Items {
		if i.Name == name {
			return nil, ErrAlreadyExists
		}
	}
	f := &File{
		fs:     dir.fs,
		parent: dir,
		cache:  proto.File{BlockSize: 4096},
	}
	dir.cache.Items = append(dir.cache.Items, proto.Item{Name: name, Type: proto.ItemFile})
	dir.open[name] = f
	f.update()
	return f, nil
}

func (dir *Dir) Delete(name string) error {
	dir.fs.mu.Lock()
	defer dir.fs.mu.Unlock()
	for i := range dir.cache.Items {
		if dir.cache.Items[i].Name == name {
			dir.cache.Items[i] = dir.cache.Items[len(dir.cache.Items)-1]
			dir.cache.Items = dir.cache.Items[:len(dir.cache.Items)-1]
			delete(dir.open, name)
			dir.update()
			return nil
		}
	}
	return ErrNotFound
}

func (dir *Dir) ReadDir() []proto.Item {
	return dir.cache.Items
}

func (dir *Dir) Attr() proto.Attr {
	return dir.cache.Attr
}

func (dir *Dir) SetAttr(attr proto.Attr) {
	dir.cache.Attr = attr
	dir.update()
}

type File struct {
	fs     *FS
	hash   []byte
	parent updater
	cache  proto.File
}

func (f *File) nodeHash() []byte {
	return f.hash
}

func (f *File) update() {
	data, _ := f.cache.Marshal(nil)
	f.hash, _ = f.fs.cas.Write(data)
	f.parent.update()
}

func (f *File) Size() uint64 {
	return f.cache.Len
}

func (f *File) Attr() proto.Attr {
	return f.cache.Attr
}

func (f *File) SetAttr(attr proto.Attr) {
	f.cache.Attr = attr
	f.update()
}

func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	f.fs.mu.Lock()
	defer f.fs.mu.Unlock()
	if off >= int64(f.cache.Len) {
		return 0, io.EOF
	}
	end := off + int64(len(p))
	n = len(p)
	if end > int64(f.cache.Len) {
		end = int64(f.cache.Len)
		n = int(f.cache.Len) - int(off)
		err = io.EOF
	}
	start := off / int64(f.cache.BlockSize)
	woff := off % int64(f.cache.BlockSize)
	for l1 := start; l1 < ((end-1)/int64(f.cache.BlockSize))+1; l1++ {
		var data []byte
		if len(f.cache.Blocks[l1]) == 0 {
			data = make([]byte, f.cache.BlockSize)
		} else {
			var err error
			data, err = f.fs.cas.Read(f.cache.Blocks[l1])
			if err != nil {
				return 0, err
			}
		}
		boff := ((l1 - start)*int64(f.cache.BlockSize)) - woff
		if boff < 0 {
			copy(p, data[-boff:])
		} else {
			copy(p[boff:], data)
		}
	}
	return
}

func (f *File) WriteAt(p []byte, off int64) (n int, err error) {
	f.fs.mu.Lock()
	defer f.fs.mu.Unlock()
	end := int64(len(p)) + off
	if uint64(end) > f.cache.Len {
		f.cache.Len = uint64(end)
	}
	neededBlocks := ((end - 1) / int64(f.cache.BlockSize)) + 1
	if int64(len(f.cache.Blocks)) < neededBlocks {
		old := f.cache.Blocks
		f.cache.Blocks = make([][]byte, neededBlocks)
		copy(f.cache.Blocks, old)
	}
	n = len(p)
	start := off / int64(f.cache.BlockSize)
	woff := off % int64(f.cache.BlockSize)
	wg := sync.WaitGroup{}
	for l1 := start; l1 < ((end-1)/int64(f.cache.BlockSize))+1; l1++ {
		wg.Add(1)
		go func(l1 int64) {
			defer wg.Done()
			var data []byte
			boff := ((l1 - start) * int64(f.cache.BlockSize)) - woff
			if (l1 == start && woff != 0) || (l1 == end/int64(f.cache.BlockSize) && end%int64(f.cache.BlockSize) != 0) {
				if len(f.cache.Blocks[l1]) == 0 {
					data = make([]byte, f.cache.BlockSize)
				} else {
					var err error
					data, err = f.fs.cas.Read(f.cache.Blocks[l1])
					if err != nil {
//						return 0, err
					}
					if uint64(len(data)) < f.cache.BlockSize {
						old := data
						data = make([]byte, f.cache.BlockSize)
						copy(data, old)
					}
				}
				if boff < 0 {
					copy(data[-boff:], p)
				} else {
					copy(data, p[boff:])
				}
			} else {
				data = p[boff : boff+int64(f.cache.BlockSize)]
			}
			var err error
			f.cache.Blocks[l1], err = f.fs.cas.Write(data)
			if err != nil {
//				return 0, err
			}
		}(l1)
	}
	wg.Wait()
	f.update()
	return
}

func (f *File) Truncate(len int64) {
	f.cache.Len = uint64(len)
	old := f.cache.Blocks
	if len > 0 {
		f.cache.Blocks = make([][]byte, (uint64(len-1)/f.cache.BlockSize)+1)
		copy(f.cache.Blocks, old)
	} else {
		f.cache.Blocks = [][]byte{}
	}
	f.update()
}
