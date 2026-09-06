//go:build nsync_spin

#include "textflag.h"

TEXT ·spinPause(SB), NOSPLIT, $0-0
	PAUSE
	RET
