#!/bin/sh
go build
for f in hello multiple; do
  (cd testdata; ../tinygocompiler "$f.go" -o "$f" && "./$f"; cd ..)
done
