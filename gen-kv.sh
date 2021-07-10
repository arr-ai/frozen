#!/bin/sh

set -e

SRCDIR=$1
OUTDIR=$2

shift 2
FILES="$@"

PACKAGE=$(basename $OUTDIR)

mkdir -p $OUTDIR

TYPE=kv.KeyValue

for f in $FILES; do
    if [ -f $OUTDIR/$f ]; then
        chmod +w $OUTDIR/$f
    fi
    ./gen-kv.pl "$PACKAGE" "$TYPE" < $SRCDIR/$f > $OUTDIR/$f
done

goimports -w -local $(head -1 go.mod | awk '{print$2}') $OUTDIR

for f in $FILES; do
    chmod -w $OUTDIR/$f
done
