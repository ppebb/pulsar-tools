package main

type OSERROR uint32

const (
	OSERROR_DSI            OSERROR = 2
	OSERROR_ISI                    = 3
	OSERROR_FLOATING_POINT         = 7
	OSERROR_FPE                    = 8
)

// Adapted from Game Structs.cs in pulsar
type GPR struct {
	Name uint32
	Gpr  uint32
}

type FPR struct {
	Name uint32
	// This field isn't present in pulsar's representation of the exception
	// handler, but seemingly there's some random bullshit stored after the
	// name??? Fuck everything.
	Padding uint32
	Fpr     float64
}

type StackFrame struct {
	SpName uint32
	Sp     uint32
	LrName uint32
	Lr     uint32
}

type ExceptionFile struct {
	Magic    uint32
	Region   uint32
	Reserved uint32
	Err      uint32
	Srr0     GPR
	Srr1     GPR
	Msr      GPR
	Cr       GPR
	Lr       GPR
	Gprs     [32]GPR
	Fprs     [32]FPR
	Fpscr    FPR
	Frames   [10]StackFrame
}

type ConfigHeader struct {
	Magic         uint32
	Version       int32
	InfoOffset    int32
	CupsOffset    int32
	BMGOffset     int32
	ModFolderName string
}

type SectionHeader struct {
	Magic   uint32
	Version uint32
	Size    uint32
}

type Info struct {
	Header               SectionHeader
	RoomKey              uint32 //transmitted to other players
	Prob100cc            uint32
	Prob150cc            uint32
	WiimmfiRegion        int32
	TrackBlocking        uint32
	HasTTTrophies        byte
	Has200cc             byte
	HasUMTs              byte
	HasFeather           byte
	HasMegaTC            byte
	CupIconCount         uint16
	ChooseNextTrackTimer byte
	ReservedSpace        [40]byte
}

type TrackV3 struct {
	Slot         byte
	MusicSlot    byte
	VariantCount int16
	Crc32        uint32
}

type Variant struct {
	Slot      byte
	MusicSlot byte
}

type CupsV3 struct {
	Header            SectionHeader
	CtsCupCount       uint16
	RegsMode          byte
	Padding           byte
	TrophyCount       [4]uint16
	TotalVariantCount int32
}

type TrackHolder struct {
	Name      string
	Main      Variant
	TrackInfo TrackV3
	Variants  []Variant
}

type CupHolder struct {
	Name   string
	Tracks []TrackHolder
}

type Config struct {
	Header  ConfigHeader
	Info    Info
	CupInfo CupsV3
	Cups    []CupHolder
}
