#!/bin/sh
go build
for f in hello; do
  (cd testdata; ../tinygocompiler "$f.go" -o "$f" && "./$f"; cd ..)
done
