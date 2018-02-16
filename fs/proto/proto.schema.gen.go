package proto

type Attr struct {
	Mode uint32
	UID  uint32
	GID  uint32
}

func (d *Attr) Size() (s uint64) {

	s += 12
	return
}
func (d *Attr) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.Mode >> 0)

		buf[1+0] = byte(d.Mode >> 8)

		buf[2+0] = byte(d.Mode >> 16)

		buf[3+0] = byte(d.Mode >> 24)

	}
	{

		buf[0+4] = byte(d.UID >> 0)

		buf[1+4] = byte(d.UID >> 8)

		buf[2+4] = byte(d.UID >> 16)

		buf[3+4] = byte(d.UID >> 24)

	}
	{

		buf[0+8] = byte(d.GID >> 0)

		buf[1+8] = byte(d.GID >> 8)

		buf[2+8] = byte(d.GID >> 16)

		buf[3+8] = byte(d.GID >> 24)

	}
	return buf[:i+12], nil
}

func (d *Attr) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		d.Mode = 0 | (uint32(buf[0+0]) << 0) | (uint32(buf[1+0]) << 8) | (uint32(buf[2+0]) << 16) | (uint32(buf[3+0]) << 24)

	}
	{

		d.UID = 0 | (uint32(buf[0+4]) << 0) | (uint32(buf[1+4]) << 8) | (uint32(buf[2+4]) << 16) | (uint32(buf[3+4]) << 24)

	}
	{

		d.GID = 0 | (uint32(buf[0+8]) << 0) | (uint32(buf[1+8]) << 8) | (uint32(buf[2+8]) << 16) | (uint32(buf[3+8]) << 24)

	}
	return i + 12, nil
}

type Item struct {
	Name string
	Type uint16
	Hash []byte
}

func (d *Item) Size() (s uint64) {

	{
		l := uint64(len(d.Name))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	{

		t := d.Type
		for t >= 0x80 {
			t >>= 7
			s++
		}
		s++

	}
	{
		l := uint64(len(d.Hash))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	return
}
func (d *Item) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{
		l := uint64(len(d.Name))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		copy(buf[i+0:], d.Name)
		i += l
	}
	{

		t := uint16(d.Type)

		for t >= 0x80 {
			buf[i+0] = byte(t) | 0x80
			t >>= 7
			i++
		}
		buf[i+0] = byte(t)
		i++

	}
	{
		l := uint64(len(d.Hash))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		copy(buf[i+0:], d.Hash)
		i += l
	}
	return buf[:i+0], nil
}

func (d *Item) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.Name = string(buf[i+0 : i+0+l])
		i += l
	}
	{

		bs := uint8(7)
		t := uint16(buf[i+0] & 0x7F)
		for buf[i+0]&0x80 == 0x80 {
			i++
			t |= uint16(buf[i+0]&0x7F) << bs
			bs += 7
		}
		i++

		d.Type = t

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Hash)) >= l {
			d.Hash = d.Hash[:l]
		} else {
			d.Hash = make([]byte, l)
		}
		copy(d.Hash, buf[i+0:])
		i += l
	}
	return i + 0, nil
}

type Directory struct {
	Items []Item
	Attr  Attr
}

func (d *Directory) Size() (s uint64) {

	{
		l := uint64(len(d.Items))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Items {

			{
				s += d.Items[k0].Size()
			}

		}

	}
	{
		s += d.Attr.Size()
	}
	return
}
func (d *Directory) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{
		l := uint64(len(d.Items))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		for k0 := range d.Items {

			{
				nbuf, err := d.Items[k0].Marshal(buf[i+0:])
				if err != nil {
					return nil, err
				}
				i += uint64(len(nbuf))
			}

		}
	}
	{
		nbuf, err := d.Attr.Marshal(buf[i+0:])
		if err != nil {
			return nil, err
		}
		i += uint64(len(nbuf))
	}
	return buf[:i+0], nil
}

func (d *Directory) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Items)) >= l {
			d.Items = d.Items[:l]
		} else {
			d.Items = make([]Item, l)
		}
		for k0 := range d.Items {

			{
				ni, err := d.Items[k0].Unmarshal(buf[i+0:])
				if err != nil {
					return 0, err
				}
				i += ni
			}

		}
	}
	{
		ni, err := d.Attr.Unmarshal(buf[i+0:])
		if err != nil {
			return 0, err
		}
		i += ni
	}
	return i + 0, nil
}

type File struct {
	Len       uint64
	Attr      Attr
	BlockSize uint64
	Blocks    [][]byte
}

func (d *File) Size() (s uint64) {

	{

		t := d.Len
		for t >= 0x80 {
			t >>= 7
			s++
		}
		s++

	}
	{
		s += d.Attr.Size()
	}
	{

		t := d.BlockSize
		for t >= 0x80 {
			t >>= 7
			s++
		}
		s++

	}
	{
		l := uint64(len(d.Blocks))

		{

			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}

		for k0 := range d.Blocks {

			{
				l := uint64(len(d.Blocks[k0]))

				{

					t := l
					for t >= 0x80 {
						t >>= 7
						s++
					}
					s++

				}
				s += l
			}

		}

	}
	return
}
func (d *File) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		t := uint64(d.Len)

		for t >= 0x80 {
			buf[i+0] = byte(t) | 0x80
			t >>= 7
			i++
		}
		buf[i+0] = byte(t)
		i++

	}
	{
		nbuf, err := d.Attr.Marshal(buf[i+0:])
		if err != nil {
			return nil, err
		}
		i += uint64(len(nbuf))
	}
	{

		t := uint64(d.BlockSize)

		for t >= 0x80 {
			buf[i+0] = byte(t) | 0x80
			t >>= 7
			i++
		}
		buf[i+0] = byte(t)
		i++

	}
	{
		l := uint64(len(d.Blocks))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		for k0 := range d.Blocks {

			{
				l := uint64(len(d.Blocks[k0]))

				{

					t := uint64(l)

					for t >= 0x80 {
						buf[i+0] = byte(t) | 0x80
						t >>= 7
						i++
					}
					buf[i+0] = byte(t)
					i++

				}
				copy(buf[i+0:], d.Blocks[k0])
				i += l
			}

		}
	}
	return buf[:i+0], nil
}

func (d *File) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{

		bs := uint8(7)
		t := uint64(buf[i+0] & 0x7F)
		for buf[i+0]&0x80 == 0x80 {
			i++
			t |= uint64(buf[i+0]&0x7F) << bs
			bs += 7
		}
		i++

		d.Len = t

	}
	{
		ni, err := d.Attr.Unmarshal(buf[i+0:])
		if err != nil {
			return 0, err
		}
		i += ni
	}
	{

		bs := uint8(7)
		t := uint64(buf[i+0] & 0x7F)
		for buf[i+0]&0x80 == 0x80 {
			i++
			t |= uint64(buf[i+0]&0x7F) << bs
			bs += 7
		}
		i++

		d.BlockSize = t

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Blocks)) >= l {
			d.Blocks = d.Blocks[:l]
		} else {
			d.Blocks = make([][]byte, l)
		}
		for k0 := range d.Blocks {

			{
				l := uint64(0)

				{

					bs := uint8(7)
					t := uint64(buf[i+0] & 0x7F)
					for buf[i+0]&0x80 == 0x80 {
						i++
						t |= uint64(buf[i+0]&0x7F) << bs
						bs += 7
					}
					i++

					l = t

				}
				if uint64(cap(d.Blocks[k0])) >= l {
					d.Blocks[k0] = d.Blocks[k0][:l]
				} else {
					d.Blocks[k0] = make([]byte, l)
				}
				copy(d.Blocks[k0], buf[i+0:])
				i += l
			}

		}
	}
	return i + 0, nil
}
