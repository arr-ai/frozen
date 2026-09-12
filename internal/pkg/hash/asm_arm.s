// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

// AES hashing not implemented for ARM
TEXT ·aeshash(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT ·aeshash32(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT ·aeshash64(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	$0, R0
	MOVW	(R0), R1
TEXT ·aeshashstr(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	$0, R0
	MOVW	(R0), R1
