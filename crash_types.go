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
