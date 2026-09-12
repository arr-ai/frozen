// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

TEXT ·aeshash(SB),NOSPLIT,$0-20
	MOVL	AX, ret+16(FP)
	RET

TEXT ·aeshashstr(SB),NOSPLIT,$0-12
	MOVL	AX, ret+8(FP)
	RET

TEXT ·aeshash32(SB),NOSPLIT,$0-12
	MOVL	AX, ret+8(FP)
	RET

TEXT ·aeshash64(SB),NOSPLIT,$0-12
	MOVL	AX, ret+8(FP)
	RET

TEXT ·return0(SB), NOSPLIT, $0
	MOVL	$0, AX
	RET
