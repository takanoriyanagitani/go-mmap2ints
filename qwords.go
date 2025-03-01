package mmap2ints

import (
	"encoding/binary"
	"iter"
	"os"

	"golang.org/x/sys/unix"
)

var PageSize int = unix.Getpagesize()

func FilenameToIntegers64(
	filename string,
) iter.Seq2[uint64, error] {
	return func(yield func(uint64, error) bool) {
		f, e := os.Open(filename)
		if nil != e {
			yield(0, e)
			return
		}
		defer f.Close()

		stat, e := f.Stat()
		if nil != e {
			yield(0, e)
			return
		}

		var size int64 = stat.Size()
		var pages int64 = size / int64(PageSize)
		var extra int64 = size - (pages * int64(PageSize))

		var fd int = int(f.Fd())

		for page := range pages {
			e := func() error {
				data, e := unix.Mmap(
					fd,
					page*(int64(PageSize)),
					PageSize,
					unix.PROT_READ,
					unix.MAP_PRIVATE,
				)
				if nil != e {
					return e
				}
				defer func() {
					e := unix.Munmap(data)
					if nil != e {
						panic(e)
					}
				}()

				if PageSize != len(data) {
					panic("invalid page size")
				}

				for ix := range PageSize / 8 {
					var start int = ix * 8
					var end int = start + 8
					var dat []byte = data[start:end]
					var u uint64 = binary.BigEndian.Uint64(dat)
					if !yield(u, nil) {
						return nil
					}
				}

				return nil
			}()

			if nil != e {
				yield(0, e)
				return
			}
		}

		if 0 == extra {
			return
		}

		data, e := unix.Mmap(
			fd,
			pages*(int64(PageSize)),
			int(extra),
			unix.PROT_READ,
			unix.MAP_PRIVATE,
		)
		if nil != e {
			yield(0, e)
			return
		}
		defer func() {
			e := unix.Munmap(data)
			if nil != e {
				panic(e)
			}
		}()

		var extraCount int64 = extra >> 3
		for ix := range extraCount {
			var start int = int(ix) * 8
			var end int = start + 8
			var dat []byte = data[start:end]
			var u uint64 = binary.BigEndian.Uint64(dat)
			if !yield(u, nil) {
				return
			}
		}
	}
}
