// check.s

#include "textflag.h"

// func nzero() int64
TEXT ·nzero(SB),NOSPLIT,$0
    MOVQ    $0, ret+0(FP)
    RET

// func nmul(a, b uint64) uint64
TEXT ·nmul(SB),NOSPLIT,$0
    MOVQ    a+0(FP), AX
    MULQ    b+8(FP) // result in EDX:EAX
    SHLQ    $32, DX // lop off high 32 bits
    SHRQ    $32, AX // lop off low 32 bits
    JCS     round // do we have a high-order fractional bit?
    ORQ     AX, DX // merge high and low 32-bit values
    MOVQ    DX, ret+16(FP)
    RET

round:
    ORQ     AX, DX // merge high and low 32-bit values
    ADDQ    $1, DX
    MOVQ    DX, ret+16(FP)
    RET

// func nmulxy(a, b uint64) (uint64, uint64)
TEXT ·nmulxy(SB),NOSPLIT,$0
    MOVQ    a+0(FP), AX
    MULQ    b+8(FP) // result in EDX:EAX
    MOVQ    DX, ret+16(FP)
    MOVQ    AX, ret+24(FP)
    RET
