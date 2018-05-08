#!/bin/sh

rm -fR dist/release/win32/split-ticket-buyer
mkdir -p dist/release/win32/split-ticket-buyer
rm -fR dist/release/win64/split-ticket-buyer
mkdir -p dist/release/win64/split-ticket-buyer
mkdir -p dist/archives

VERSION=`grep -oP "Version\s+ = \"\K[^\"]+(?=\")" pkg/version.go`

echo "Building binaries $VERSION..."

echo "Building CLI buyer (win32)"
env GOOS=windows GOARCH=386 \
    go build \
    -v \
    -o dist/release/win32/split-ticket-buyer/splitticketbuyer \
    cmd/splitticketbuyer/*.go
if [[ $? != 0 ]] ; then exit 1 ; fi

echo "Building GUI buyer (win32)"
env GOOS=windows GOARCH=386 \
    go build \
    -v \
    -o dist/release/win32/split-ticket-buyer/splitticketbuyergui \
    cmd/splitticketbuyergui/*.go
if [[ $? != 0 ]] ; then exit 1 ; fi

cp docs/release-readme.md dist/release/win32/split-ticket-buyer/README.md

ZIPFILE="splitticketbuyer-win32-$VERSION.tar.gz"

rm -f dist/archives/$ZIPFILE

cd dist/release/win32 && tar -czf ../../archives/$ZIPFILE split-ticket-buyer

echo "Built win32 binaries $VERSION"


### win64

echo "Building CLI buyer (win64)"
env GOOS=windows GOARCH=amd64 \
    go build \
    -v \
    -o dist/release/win64/split-ticket-buyer/splitticketbuyer \
    cmd/splitticketbuyer/*.go
if [[ $? != 0 ]] ; then exit 1 ; fi

echo "Building GUI buyer (win64)"
env GOOS=windows GOARCH=amd64 \
    go build \
    -v \
    -o dist/release/win64/split-ticket-buyer/splitticketbuyergui \
    cmd/splitticketbuyergui/*.go
if [[ $? != 0 ]] ; then exit 1 ; fi

cp docs/release-readme.md dist/release/win64/split-ticket-buyer/README.md

ZIPFILE="splitticketbuyer-win64-$VERSION.tar.gz"

rm -f dist/archives/$ZIPFILE

cd dist/release/win64 && tar -czf ../../archives/$ZIPFILE split-ticket-buyer

echo "Built win64 binaries $VERSION"
