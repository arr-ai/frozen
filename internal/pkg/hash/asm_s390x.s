// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

// AES hashing not implemented for s390x
TEXT ·aeshash(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	(R0), R15
TEXT ·aeshash32(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	(R0), R15
TEXT ·aeshash64(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	(R0), R15
TEXT ·aeshashstr(SB),NOSPLIT|NOFRAME,$0-0
	MOVW	(R0), R15
