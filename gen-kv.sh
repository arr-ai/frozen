#!/bin/sh

set -e

SRCDIR=$1
OUTDIR=$2

shift 2
FILES="$@"

PACKAGE=$(basename $OUTDIR)

mkdir -p $OUTDIR

for f in $FILES; do
    perl -ne "
        BEGIN {
            \$PRINT = 1;
        }

        s{^package \w+$}{package $PACKAGE\nimport \"github.com/arr-ai/frozen/pkg/kv\"};
        s/\binterface{}/kv.KeyValue/g;
        s/\bactualInterface\b/interface{}/g;
        if ('$PACKAGE' eq 'kvi') {
        } elsif ('$PACKAGE' eq 'kvt') {
            s/\biterator\.(Iterator|Empty)\b/kvi.\1/g;
            s/\bvalue.Equal\b/KeyEqual/g;
            s{(^package \w*)}{\1\nimport \"github.com/arr-ai/frozen/pkg/kv\"};
         	s{\"github.com/arr-ai/frozen/internal/iterator\"}{\"github.com/arr-ai/frozen/internal/iterator/kvi\"};
        }

        if (m{^\\s*// SUBST (KeyValue|%): (.*) => (.*)\$}) {
            \$FROM = \$2;
            \$TO = \$3;
            \$TO =~ s/%/KeyValue/ if \$1 eq '%';
        } elsif (m{^(.*)// SUBST (KeyValue|%): (.*) => (.*)\$}) {
            s/\$2/\$1\$3/g;
        } elsif (m{^\\s*// !SUBST\$}) {
            undef \$FROM;
            undef \$TO;
        } elsif (m{^\\s*// ELIDE (KeyValue|%)\$}) {
            \$PRINT = 0;
        } elsif (m{^\\s*// !ELIDE\$}) {
            \$PRINT = 1;
        } else {
            s/\$FROM/\$TO/g if \$FROM;
            print if \$PRINT;
        }
    " $SRCDIR/$f > $OUTDIR/$f
done

goimports -w $OUTDIR
