// time_windows_amd64.s

#include "textflag.h"

// func nmul(a, b uint64) uint64
TEXT ·nmul(SB),NOSPLIT,$0
    MOVQ    a+0(FP), AX
    MULQ    b+8(FP) // result in EDX:EAX
    SHLQ    $32, DX // lop off high 32 bits
    SHRQ    $32, AX // lop off low 32 bits
    ORQ     AX, DX // merge high and low 32-bit values
    MOVQ    DX, ret+16(FP)
    RET
