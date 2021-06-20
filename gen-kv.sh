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
    perl -ne "
        BEGIN {
            \$PRINT = 1;
        }

        s{^package \w+$}{package $PACKAGE\nimport \"github.com/arr-ai/frozen/pkg/kv\"};
        s/\binterface{}/$TYPE/g;
        s/\bactualInterface\b/interface{}/g;
        if ('$PACKAGE' eq 'kvi') {
        } elsif ('$PACKAGE' eq 'kvt') {
            s/\biterator\.(Iterator|Empty)\b/kvi.\1/g;
            s/\bvalue.Equal\b/KeyEqual/g;
            s{(^package \w*)}{\1\nimport \"github.com/arr-ai/frozen/pkg/kv\"};
         	s{\"github.com/arr-ai/frozen/internal/iterator\"}{\"github.com/arr-ai/frozen/internal/iterator/kvi\"};
        }

        if (m{^\\s*// SUBST ($TYPE|%): (.*) => (.*)\$}) {
            print STDERR \"????\n\";
            \$FROM = \$2;
            \$TO = \$3;
            \$TO =~ s/%/$TYPE/ if \$1 eq '%';
        } elsif (m{^(.*?)\\s*// SUBST ($TYPE|%): (.*) => (.*)\$}) {
            \$_ = \$1 . \"\\n\";
            \$from = \$3;
            \$to = \$4;
            \$to =~ s/%/$TYPE/ if \$2 eq '%';
            s/\$from/\$to/g;
            print;
        } elsif (m{^\\s*// !SUBST\$}) {
            undef \$FROM;
            undef \$TO;
        } elsif (m{^\\s*// ELIDE ($TYPE|%)\$}) {
            \$PRINT = 0;
        } elsif (m{^\\s*// !ELIDE\$}) {
            \$PRINT = 1;
        } else {
            s/\$FROM/\$TO/g if \$FROM;
            print if \$PRINT;
        }
    " $SRCDIR/$f > $OUTDIR/$f
done

goimports -w -local $(head -1 go.mod | awk '{print$2}') $OUTDIR
