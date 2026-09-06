#include "textflag.h"

TEXT ·acquire(SB), NOSPLIT, $0-9
	MOVD ptr+0(FP), R0
	MOVW $1, R2
#ifdef GOARM64_LSE
	MOVW $0, R1
	CASALW R1, (R0), R2
	CMPW $0, R1
	CSET EQ, R3
	MOVB R3, ret+8(FP)
	RET
#else
retry_acquire:
	LDAXRW (R0), R1
	CBNZW R1, failed
	STLXRW R2, (R0), R3
	CBNZW R3, retry_acquire
	MOVD $1, R3
	MOVB R3, ret+8(FP)
	RET
failed:
	MOVB ZR, ret+8(FP)
	RET
#endif

TEXT ·release(SB), NOSPLIT, $0-9
	MOVD ptr+0(FP), R0
	MOVW $0, R2
#ifdef GOARM64_LSE
	MOVW $1, R1
	CASALW R1, (R0), R2
	CMPW $1, R1
	CSET EQ, R3
	MOVB R3, ret+8(FP)
	RET
#else
retry_release:
	LDAXRW (R0), R1
	CMPW $1, R1
	BNE failed_release
	STLXRW R2, (R0), R3
	CBNZW R3, retry_release
	MOVD $1, R3
	MOVB R3, ret+8(FP)
	RET
failed_release:
	MOVB ZR, ret+8(FP)
	RET
#endif
