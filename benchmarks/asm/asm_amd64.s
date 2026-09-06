#include "textflag.h"

TEXT ·acquire(SB), NOSPLIT, $0-9
	MOVQ ptr+0(FP), CX
	XORL AX, AX
	MOVL $1, DX
	LOCK
	CMPXCHGL DX, (CX)
	SETEQ ret+8(FP)
	RET

TEXT ·release(SB), NOSPLIT, $0-9
	MOVQ ptr+0(FP), CX
	MOVL $1, AX
	XORL DX, DX
	LOCK
	CMPXCHGL DX, (CX)
	SETEQ ret+8(FP)
	RET
