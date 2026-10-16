package uab

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type secDef struct {
	name string
	typ  elf.SectionType
	data []byte
}

type segRange struct {
	off    uint64
	filesz uint64
}

func alignUp(v, a uint64) uint64 {
	if a <= 1 {
		return v
	}
	if r := v % a; r != 0 {
		v += a - r
	}
	return v
}

func buildELF(t *testing.T, order binary.ByteOrder, machine elf.Machine,
	sections []secDef, trailing []byte, segCount int,
	segments func(offsets map[string]uint64, fileSize uint64) []segRange,
) []byte {
	t.Helper()

	var dataByte byte
	switch order {
	case binary.LittleEndian:
		dataByte = byte(elf.ELFDATA2LSB)
	case binary.BigEndian:
		dataByte = byte(elf.ELFDATA2MSB)
	default:
		t.Fatalf("unsupported byte order %v", order)
	}

	shstr := []byte{0}
	nameOffsets := make([]uint32, len(sections))
	shstrndx := -1
	for i, s := range sections {
		if s.name == ".shstrtab" {
			shstrndx = i + 1
		}
		nameOffsets[i] = uint32(len(shstr))
		shstr = append(shstr, s.name...)
		shstr = append(shstr, 0)
	}
	if shstrndx < 0 {
		t.Fatal("sections must include .shstrtab")
	}

	phOff := uint64(0)
	if segCount > 0 {
		phOff = elf64EhdrSize
	}
	cur := alignUp(elf64EhdrSize+uint64(segCount)*elf64ProgHdrSize, 8)

	offsets := map[string]uint64{}
	for _, s := range sections {
		offsets[s.name] = cur
		n := uint64(len(s.data))
		if s.name == ".shstrtab" {
			n = uint64(len(shstr))
		}
		cur += n
	}

	shtOff := alignUp(cur, 8)
	realCount := len(sections) + 1
	fileSize := shtOff + uint64(realCount)*elf64SectionHdrSize + uint64(len(trailing))

	var segs []segRange
	if segments != nil {
		segs = segments(offsets, fileSize)
		if len(segs) != segCount {
			t.Fatalf("segment count mismatch: got %d want %d", len(segs), segCount)
		}
	}

	buf := make([]byte, fileSize)

	hdr := elf.Header64{
		Ident:     [16]byte{0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), dataByte, byte(elf.EV_CURRENT)},
		Type:      uint16(elf.ET_EXEC),
		Machine:   uint16(machine),
		Version:   uint32(elf.EV_CURRENT),
		Phoff:     phOff,
		Shoff:     shtOff,
		Ehsize:    elf64EhdrSize,
		Phentsize: elf64ProgHdrSize,
		Phnum:     uint16(segCount),
		Shentsize: elf64SectionHdrSize,
		Shnum:     uint16(realCount),
		Shstrndx:  uint16(shstrndx),
	}

	var hb bytes.Buffer
	require.NoError(t, binary.Write(&hb, order, &hdr))
	copy(buf, hb.Bytes())

	for i, s := range segs {
		var pb bytes.Buffer
		require.NoError(t, binary.Write(&pb, order, uint32(elf.PT_LOAD)))
		require.NoError(t, binary.Write(&pb, order, uint32(elf.PF_R)))
		require.NoError(t, binary.Write(&pb, order, s.off))
		require.NoError(t, binary.Write(&pb, order, s.off)) // p_vaddr
		require.NoError(t, binary.Write(&pb, order, s.off)) // p_paddr
		require.NoError(t, binary.Write(&pb, order, s.filesz))
		require.NoError(t, binary.Write(&pb, order, s.filesz))  // p_memsz
		require.NoError(t, binary.Write(&pb, order, uint64(1))) // p_align
		copy(buf[phOff+uint64(i)*elf64ProgHdrSize:], pb.Bytes())
	}

	for _, s := range sections {
		if s.name == ".shstrtab" {
			copy(buf[offsets[s.name]:], shstr)
		} else {
			copy(buf[offsets[s.name]:], s.data)
		}
	}

	if len(trailing) > 0 {
		copy(buf[fileSize-uint64(len(trailing)):], trailing)
	}

	for i, s := range sections {
		sh := elf.Section64{
			Name:      nameOffsets[i],
			Type:      uint32(s.typ),
			Off:       offsets[s.name],
			Size:      uint64(len(s.data)),
			Addralign: 1,
		}

		if s.name == ".shstrtab" {
			sh.Type = uint32(elf.SHT_STRTAB)
			sh.Size = uint64(len(shstr))
		}

		var sb bytes.Buffer
		require.NoError(t, binary.Write(&sb, order, &sh))
		copy(buf[shtOff+uint64(i+1)*elf64SectionHdrSize:], sb.Bytes())
	}

	return buf
}

func writeTempELF(t *testing.T, order binary.ByteOrder, machine elf.Machine,
	sections []secDef, trailing []byte, segCount int,
	segments func(offsets map[string]uint64, fileSize uint64) []segRange,
) string {
	t.Helper()
	data := buildELF(t, order, machine, sections, trailing, segCount, segments)
	path := filepath.Join(t.TempDir(), "test.elf")
	require.NoError(t, os.WriteFile(path, data, 0o644))
	return path
}

func signAction(t *testing.T, path string, update bool, payload string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	require.NoError(t, writeELFSection(f, signSectionName,
		bytes.NewReader([]byte(payload)), uint64(len(payload)), update))
}

func requireSectionData(t *testing.T, path string, want map[string][]byte) {
	t.Helper()
	ef, err := elf.Open(path)
	require.NoError(t, err)
	defer func() { _ = ef.Close() }()
	for name, data := range want {
		s := ef.Section(name)
		require.NotNil(t, s, "section %q missing", name)
		got, err := s.Data()
		require.NoError(t, err)
		require.Equal(t, data, got, "section %q", name)
	}
}

func elfFileSize(t *testing.T, path string) int64 {
	t.Helper()
	fi, err := os.Stat(path)
	require.NoError(t, err)
	return fi.Size()
}

func TestWriteELFSectionAcrossArchitectures(t *testing.T) {
	cases := []struct {
		name    string
		order   binary.ByteOrder
		machine elf.Machine
	}{
		{"x86_64-little", binary.LittleEndian, elf.EM_X86_64},
		{"aarch64-little", binary.LittleEndian, elf.EM_AARCH64},
		{"riscv-little", binary.LittleEndian, elf.EM_RISCV},
		{"s390x-big", binary.BigEndian, elf.EM_S390},
		{"ppc64-big", binary.BigEndian, elf.EM_PPC64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempELF(t, tc.order, tc.machine,
				[]secDef{{name: ".shstrtab", typ: elf.SHT_STRTAB}}, nil, 0, nil)

			payload := []byte("first-signature")
			signAction(t, path, false, string(payload))

			assertELFSection(t, path, tc.order, tc.machine, payload, 3)

			updated := []byte("second-signature")
			signAction(t, path, true, string(updated))
			assertELFSection(t, path, tc.order, tc.machine, updated, 3)

			f, err := os.OpenFile(path, os.O_RDWR, 0)
			require.NoError(t, err)
			defer func() { _ = f.Close() }()
			require.Error(t, writeELFSection(f, "not.exist", bytes.NewReader(payload), uint64(len(payload)), true))
		})
	}
}

func TestWriteELFSectionOddLayouts(t *testing.T) {
	const oldSign = "old-signature"

	t.Run("add-when-shstr-not-last", func(t *testing.T) {
		path := writeTempELF(t, binary.LittleEndian, elf.EM_X86_64, []secDef{
			{name: ".shstrtab", typ: elf.SHT_STRTAB},
			{name: ".comment", typ: elf.SHT_PROGBITS, data: []byte("comment-data")},
		}, nil, 0, nil)

		signAction(t, path, false, "sig-payload")

		requireSectionData(t, path, map[string][]byte{
			".comment":      []byte("comment-data"),
			signSectionName: []byte("sig-payload"),
		})
	})

	t.Run("update-compacts-when-shstr-after-sign", func(t *testing.T) {
		path := writeTempELF(t, binary.LittleEndian, elf.EM_X86_64, []secDef{
			{name: signSectionName, typ: elf.SHT_PROGBITS, data: []byte(oldSign)},
			{name: ".shstrtab", typ: elf.SHT_STRTAB},
		}, nil, 0, nil)
		before := elfFileSize(t, path)

		signAction(t, path, true, "new-signature")

		requireSectionData(t, path, map[string][]byte{signSectionName: []byte("new-signature")})
		require.LessOrEqual(t, elfFileSize(t, path), before)
	})

	t.Run("update-appends-when-intruder-after-sign", func(t *testing.T) {
		path := writeTempELF(t, binary.LittleEndian, elf.EM_X86_64, []secDef{
			{name: signSectionName, typ: elf.SHT_PROGBITS, data: []byte(oldSign)},
			{name: ".comment", typ: elf.SHT_PROGBITS, data: []byte("comment-data")},
			{name: ".shstrtab", typ: elf.SHT_STRTAB},
		}, nil, 0, nil)

		signAction(t, path, true, "new")

		requireSectionData(t, path, map[string][]byte{
			".comment":      []byte("comment-data"),
			signSectionName: []byte("new"),
		})
	})
}

func TestWriteELFSectionPreservesTrailingData(t *testing.T) {
	trailing := []byte("APPENDED-TRAILER-BYTES")
	orig := buildELF(t, binary.LittleEndian, elf.EM_X86_64, []secDef{
		{name: ".shstrtab", typ: elf.SHT_STRTAB},
		{name: signSectionName, typ: elf.SHT_PROGBITS, data: []byte("old-signature")},
	}, trailing, 0, nil)

	path := filepath.Join(t.TempDir(), "test.elf")
	require.NoError(t, os.WriteFile(path, orig, 0o644))

	signAction(t, path, true, "new-signature")

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, trailing, got[len(orig)-len(trailing):len(orig)], "trailing bytes must be preserved")
	requireSectionData(t, path, map[string][]byte{signSectionName: []byte("new-signature")})
}

func TestWriteELFSectionPreservesSegmentData(t *testing.T) {
	var segStart uint64
	orig := buildELF(t, binary.LittleEndian, elf.EM_X86_64, []secDef{
		{name: ".shstrtab", typ: elf.SHT_STRTAB},
		{name: signSectionName, typ: elf.SHT_PROGBITS, data: []byte("old-signature")},
	}, nil, 1, func(offsets map[string]uint64, fileSize uint64) []segRange {
		segStart = offsets[signSectionName]
		return []segRange{{off: segStart, filesz: fileSize - segStart}}
	})

	path := filepath.Join(t.TempDir(), "test.elf")
	require.NoError(t, os.WriteFile(path, orig, 0o644))

	signAction(t, path, true, "new-signature")

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, orig[segStart:], got[segStart:uint64(len(orig))], "segment-covered bytes must be preserved")
	requireSectionData(t, path, map[string][]byte{signSectionName: []byte("new-signature")})
}

func assertELFSection(t *testing.T, path string, order binary.ByteOrder, machine elf.Machine, want []byte, wantSections int) {
	t.Helper()

	ef, err := elf.Open(path)
	require.NoError(t, err)
	defer func() { _ = ef.Close() }()

	require.Equal(t, order, ef.ByteOrder)
	require.Equal(t, machine, ef.Machine)
	require.Len(t, ef.Sections, wantSections)

	section := ef.Section(signSectionName)
	require.NotNil(t, section)
	got, err := section.Data()
	require.NoError(t, err)
	require.Equal(t, want, got)
}
