#!/usr/bin/env sh
set -eu

version=${1:?usage: scripts/build-release-assets.sh VERSION}
output_dir=${OUTPUT_DIR:-dist}
mkdir -p "$output_dir"
rm -f "$output_dir"/gitu_* "$output_dir"/checksums.txt

for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
  os=${target%/*}
  arch=${target#*/}
  extension=''
  [ "$os" = windows ] && extension=.exe
  name="gitu_${version}_${os}_${arch}${extension}"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags='-s -w' -o "$output_dir/$name" ./cmd/gitu
done

(cd "$output_dir" && shasum -a 256 gitu_* > checksums.txt)
