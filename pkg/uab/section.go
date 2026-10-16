package uab

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	elf64OffShoff       = 0x28
	elf64OffShentsize   = 0x3a
	elf64OffShnum       = 0x3c
	elf64OffShstrndx    = 0x3e
	elf64SectionHdrSize = 64
	elf64ProgHdrSize    = 56
	elf64EhdrSize       = 64
	elf64ShdrOffOffset  = 24
	elf64ShdrOffSize    = 32
)

func writeELFSection(f *os.File, name string, payload io.Reader, size uint64, update bool) error {
	ef, err := elf.NewFile(f)
	if err != nil {
		return fmt.Errorf("parse elf: %w", err)
	}

	if ef.Class != elf.ELFCLASS64 {
		return fmt.Errorf("unsupported ELF class: %s (only 64-bit ELF is supported)", ef.Class)
	}
	order := ef.ByteOrder

	shentsize, err := readUint16At(f, order, elf64OffShentsize)
	if err != nil {
		return fmt.Errorf("read e_shentsize: %w", err)
	}

	if shentsize != elf64SectionHdrSize {
		return fmt.Errorf("unsupported section header entry size: %d (only 64-bit ELF is supported)", shentsize)
	}

	existing := -1
	for i, s := range ef.Sections {
		if s.Name == name {
			existing = i
			break
		}
	}

	if existing >= 0 && !update {
		return fmt.Errorf("section %s already exists", name)
	}

	if existing < 0 && update {
		return fmt.Errorf("no section %s to update", name)
	}

	if existing >= 0 {
		return updateELFSection(f, ef, order, existing, payload, size)
	}

	return addELFSection(f, ef, order, name, payload, size)
}

func updateELFSection(f *os.File, ef *elf.File, order binary.ByteOrder, index int, payload io.Reader, size uint64) error {
	rawSHT, err := readRawSHT(f, order, len(ef.Sections))
	if err != nil {
		return err
	}
	sign := ef.Sections[index]

	ok, err := tailCanBeRewritten(f, ef, order, int64(sign.Offset), map[int]bool{index: true})
	if err != nil {
		return err
	}

	if ok {
		if err := f.Truncate(int64(sign.Offset)); err != nil {
			return fmt.Errorf("truncate file: %w", err)
		}

		if err := writeAt(f, int64(sign.Offset), payload, size); err != nil {
			return fmt.Errorf("write payload: %w", err)
		}

		patchSectionEntry(order, rawSHT, index, int64(sign.Offset), size)
		return appendSHT(f, order, rawSHT)
	}

	shstrndx, err := sectionStringTableIndex(f, ef, order)
	if err != nil {
		return err
	}

	if shstrndx != index {
		shstr := ef.Sections[shstrndx]
		tailStart := int64(sign.Offset)
		if int64(shstr.Offset) < tailStart {
			tailStart = int64(shstr.Offset)
		}

		ok, err := tailCanBeRewritten(f, ef, order, tailStart, map[int]bool{index: true, shstrndx: true})
		if err != nil {
			return err
		}

		if ok {
			shstrData, err := shstr.Data()
			if err != nil {
				return fmt.Errorf("read .shstrtab: %w", err)
			}

			return compactTail(f, ef, order, rawSHT, tailStart, shstrndx, shstrData, index, payload, size)
		}
	}

	offset, err := appendAlignedFile(f, sign.Addralign)
	if err != nil {
		return err
	}

	if err := writeAt(f, offset, payload, size); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}

	patchSectionEntry(order, rawSHT, index, offset, size)
	return appendSHT(f, order, rawSHT)
}

func addELFSection(f *os.File, ef *elf.File, order binary.ByteOrder, name string, payload io.Reader, size uint64) error {
	shstrndx, err := sectionStringTableIndex(f, ef, order)
	if err != nil {
		return err
	}

	rawSHT, err := readRawSHT(f, order, len(ef.Sections))
	if err != nil {
		return err
	}

	shstr, err := ef.Sections[shstrndx].Data()
	if err != nil {
		return fmt.Errorf("read .shstrtab: %w", err)
	}

	signNameOffset := uint32(len(shstr))
	shstr = append(shstr, name...)
	shstr = append(shstr, 0)

	old := ef.Sections[shstrndx]
	shstrOff := int64(old.Offset)
	reuse, err := tailCanBeRewritten(f, ef, order, shstrOff, map[int]bool{shstrndx: true})
	if err != nil {
		return err
	}

	if reuse {
		if err := f.Truncate(shstrOff); err != nil {
			return fmt.Errorf("truncate file: %w", err)
		}
	} else {
		fi, err := f.Stat()
		if err != nil {
			return fmt.Errorf("stat file: %w", err)
		}

		shstrOff = alignOffset(fi.Size(), old.Addralign)
	}

	if err := writeAt(f, shstrOff, bytes.NewReader(shstr), uint64(len(shstr))); err != nil {
		return fmt.Errorf("write .shstrtab: %w", err)
	}

	patchSectionEntry(order, rawSHT, shstrndx, shstrOff, uint64(len(shstr)))

	signOff := alignOffset(shstrOff+int64(len(shstr)), 1)
	if err := writeAt(f, signOff, payload, size); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}

	var entry bytes.Buffer
	if err := binary.Write(&entry, order, elf.Section64{
		Name:      signNameOffset,
		Type:      uint32(elf.SHT_PROGBITS),
		Off:       uint64(signOff),
		Size:      size,
		Addralign: 1,
	}); err != nil {
		return fmt.Errorf("encode section entry: %w", err)
	}

	rawSHT = append(rawSHT, entry.Bytes()...)

	shnum := len(ef.Sections) + 1
	shnumField := uint16(shnum)
	if shnum >= int(elf.SHN_LORESERVE) {
		patchSectionEntry(order, rawSHT, 0, 0, uint64(shnum))
		shnumField = 0
	}

	if err := appendSHT(f, order, rawSHT); err != nil {
		return err
	}

	if err := writeUint16At(f, order, elf64OffShnum, shnumField); err != nil {
		return fmt.Errorf("write e_shnum: %w", err)
	}

	return nil
}

func compactTail(f *os.File, ef *elf.File, order binary.ByteOrder, rawSHT []byte,
	tailStart int64, shstrndx int, shstrData []byte, signIdx int, payload io.Reader, size uint64,
) error {
	shstr := ef.Sections[shstrndx]
	sign := ef.Sections[signIdx]

	shstrOff := alignOffset(tailStart, shstr.Addralign)
	if err := writeAt(f, shstrOff, bytes.NewReader(shstrData), uint64(len(shstrData))); err != nil {
		return fmt.Errorf("write .shstrtab: %w", err)
	}

	signOff := alignOffset(shstrOff+int64(len(shstrData)), sign.Addralign)
	if err := writeAt(f, signOff, payload, size); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}

	if err := f.Truncate(signOff + int64(size)); err != nil {
		return fmt.Errorf("truncate file: %w", err)
	}

	patchSectionEntry(order, rawSHT, shstrndx, shstrOff, uint64(len(shstrData)))
	patchSectionEntry(order, rawSHT, signIdx, signOff, size)
	return appendSHT(f, order, rawSHT)
}

func tailCanBeRewritten(f *os.File, ef *elf.File, order binary.ByteOrder, tailStart int64, allowed map[int]bool) (bool, error) {
	if tailStart < elf64EhdrSize {
		return false, nil
	}

	for i, s := range ef.Sections {
		if allowed[i] || s.Size == 0 || s.Type == elf.SHT_NULL || s.Type == elf.SHT_NOBITS {
			continue
		}

		if int64(s.Offset+s.Size) > tailStart {
			return false, nil
		}
	}

	for _, p := range ef.Progs {
		if p.Filesz == 0 {
			continue
		}

		if int64(p.Off+p.Filesz) > tailStart {
			return false, nil
		}
	}

	fi, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}

	end := int64(0)
	upd := func(v uint64) {
		if int64(v) > end {
			end = int64(v)
		}
	}

	for _, s := range ef.Sections {
		if s.Type == elf.SHT_NULL || s.Type == elf.SHT_NOBITS || s.Size == 0 {
			continue
		}

		upd(s.Offset + s.Size)
	}

	for _, p := range ef.Progs {
		if p.Filesz > 0 {
			upd(p.Off + p.Filesz)
		}
	}

	shoff, err := readUint64At(f, order, elf64OffShoff)
	if err != nil {
		return false, fmt.Errorf("read e_shoff: %w", err)
	}
	upd(shoff + uint64(len(ef.Sections)*elf64SectionHdrSize))

	return fi.Size() <= end, nil
}

func sectionStringTableIndex(f *os.File, ef *elf.File, order binary.ByteOrder) (int, error) {
	index, err := readUint16At(f, order, elf64OffShstrndx)
	if err != nil {
		return 0, fmt.Errorf("read e_shstrndx: %w", err)
	}

	idx := int(index)
	if idx == int(elf.SHN_XINDEX) {
		if len(ef.Sections) == 0 {
			return 0, fmt.Errorf("missing section header table")
		}
		idx = int(ef.Sections[0].Link)
	}

	if idx <= 0 || idx >= len(ef.Sections) {
		return 0, fmt.Errorf("invalid section string table index: %d", idx)
	}

	return idx, nil
}

func patchSectionEntry(order binary.ByteOrder, rawSHT []byte, index int, offset int64, size uint64) {
	entry := rawSHT[index*elf64SectionHdrSize : (index+1)*elf64SectionHdrSize]
	order.PutUint64(entry[elf64ShdrOffOffset:], uint64(offset))
	order.PutUint64(entry[elf64ShdrOffSize:], size)
}

func appendSHT(f *os.File, order binary.ByteOrder, rawSHT []byte) error {
	shtOffset, err := appendAligned(f, rawSHT, 8)
	if err != nil {
		return err
	}

	if err := writeUint64At(f, order, elf64OffShoff, uint64(shtOffset)); err != nil {
		return fmt.Errorf("write e_shoff: %w", err)
	}

	return nil
}

func appendAlignedFile(f *os.File, align uint64) (int64, error) {
	fi, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat file: %w", err)
	}

	return alignOffset(fi.Size(), align), nil
}

func readRawSHT(f *os.File, order binary.ByteOrder, count int) ([]byte, error) {
	shoff, err := readUint64At(f, order, elf64OffShoff)
	if err != nil {
		return nil, fmt.Errorf("read e_shoff: %w", err)
	}
	raw := make([]byte, count*elf64SectionHdrSize)
	if _, err := f.ReadAt(raw, int64(shoff)); err != nil {
		return nil, fmt.Errorf("read SHT: %w", err)
	}
	return raw, nil
}

func writeAt(f *os.File, offset int64, r io.Reader, size uint64) error {
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	_, err := io.CopyN(f, r, int64(size))
	return err
}

func appendAligned(f *os.File, data []byte, align uint64) (int64, error) {
	fi, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat file: %w", err)
	}

	offset := alignOffset(fi.Size(), align)
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}

	if _, err := f.Write(data); err != nil {
		return 0, err
	}

	return offset, nil
}

func alignOffset(offset int64, align uint64) int64 {
	if align <= 1 {
		return offset
	}

	a := int64(align)
	if rem := offset % a; rem != 0 {
		offset += a - rem
	}

	return offset
}

func readUint16At(f *os.File, order binary.ByteOrder, offset int64) (uint16, error) {
	var buf [2]byte
	if _, err := f.ReadAt(buf[:], offset); err != nil {
		return 0, err
	}

	return order.Uint16(buf[:]), nil
}

func readUint64At(f *os.File, order binary.ByteOrder, offset int64) (uint64, error) {
	var buf [8]byte
	if _, err := f.ReadAt(buf[:], offset); err != nil {
		return 0, err
	}

	return order.Uint64(buf[:]), nil
}

func writeUint16At(f *os.File, order binary.ByteOrder, offset int64, v uint16) error {
	var buf [2]byte
	order.PutUint16(buf[:], v)
	_, err := f.WriteAt(buf[:], offset)
	return err
}

func writeUint64At(f *os.File, order binary.ByteOrder, offset int64, v uint64) error {
	var buf [8]byte
	order.PutUint64(buf[:], v)
	_, err := f.WriteAt(buf[:], offset)
	return err
}
