package net

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	pe_signature_offset = 0x3c
	pe32                = 0x10b
	pe32p               = 0x20b
)

type coff_file_header struct {
	Magic                []byte `length:"4"`
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
	OptionalHeader       optional_header
	Sections             []section_table `length:"NumberOfSections"`
}

func (coff *coff_file_header) Validate() error {
	if bytes.Compare(coff.Magic, pe_magic) != 0 {
		return errors.New(fmt.Sprintf("PE Magic mismatch: %v", coff.Magic))
	}
	return nil
}

func (coff *coff_file_header) VirtualToFileOffset(addr uint32) uint32 {
	off := uint32(0)
	sec := 0
	for addr > coff.Sections[sec+1].VirtualAddress && sec < len(coff.Sections)-2 {
		sec++
	}
	off = addr - coff.Sections[sec].VirtualAddress + coff.Sections[sec].PointerToRawData
	return off
}

type optional_header struct {
	Magic                       uint16
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	BaseOfData                  uint32 `if:"Magic,0x10b"`
	ImageBase                   uint32 `if:"Magic,0x10b"`
	ImageBase64                 uint64 `if:"Magic,0x20b"`
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint32 `if:"Magic,0x10b"`
	SizeOfStackCommit           uint32 `if:"Magic,0x10b"`
	SizeOfHeapReserve           uint32 `if:"Magic,0x10b"`
	SizeOfHeapCommit            uint32 `if:"Magic,0x10b"`
	SizeOfStackReserve64        uint64 `if:"Magic,0x20b"`
	SizeOfStackCommit64         uint64 `if:"Magic,0x20b"`
	SizeOfHeapReserve64         uint64 `if:"Magic,0x20b"`
	SizeOfHeapCommit64          uint64 `if:"Magic,0x20b"`
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
	RVAS                        []image_data_directory `length:"NumberOfRvaAndSizes"`
}

func (o *optional_header) Validate() error {
	if o.Magic != pe32 && o.Magic != pe32p {
		return errors.New(fmt.Sprintf("Unkown optional header magic: %x", o.Magic))
	}
	if len(o.RVAS) != 16 || o.RVAS[14].VirtualAddress == 0 || o.RVAS[14].Size == 0 {
		return ErrNotAssembly
	}
	return nil
}

type image_data_directory struct {
	VirtualAddress uint32
	Size           uint32
}

type section_table struct {
	Name                 string `length:"8"`
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLineNumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

func (s *section_table) Validate() error {
	if s.Name[0] != '.' {
		return errors.New(fmt.Sprintf("This does not appear to be a valid section header: %#v", s))
	}
	return nil
}

type image_cor20 struct {
	Size         uint32
	MajorVersion uint16
	MinorVersion uint16
	MetaData     image_data_directory
	Flags        uint32
}

func (cor20 *image_cor20) Validate() error {
	if cor20.MetaData.VirtualAddress == 0 || cor20.MetaData.Size == 0 {
		return ErrNotAssembly
	}
	return nil
}

func (s *section_table) String() string {
	return fmt.Sprintf("Name: %s, VirtualAddress: %d, PointerToRawData: %d", s.Name, s.VirtualAddress, s.PointerToRawData)
}

var (
	pe_magic = []byte("PE\u0000\u0000")
)
