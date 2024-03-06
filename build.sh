#!/bin/sh
set -e

# GO build options
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# Checking for build programs
if ! command -v go 2>&1 /dev/null; then
    echo "Unable to find go"
    exit 1
fi
if ! command -v make 2>&1 /dev/null; then
    echo "Unable to find make"
    exit 1
fi
if ! command -v autoconf 2>&1 /dev/null; then
    echo "Unable to find autoconf"
    exit 1
fi
if ! command -v patchelf 2>&1 /dev/null; then
    echo "Unable to find patchelf"
    exit 1
fi
if ! command -v sed 2>&1 /dev/null; then
    echo "Unable to find sed"
    exit 1
fi
if ! command -v minisign 2>&1 /dev/null; then
    echo "Unable to find minisign"
    exit 1
fi

if [ ! -f ./verifyConfig.go ]; then
    echo "Unable to find verifyConfig.go"
    exit 1
fi

if [ ! -f ./graphic.bvg ]; then
    echo "Unable to find graphic.bvg"
    exit 1
fi

mkdir -p fsverify_root/bin
mkdir -p fsverify_root/share
mkdir -p fsverify_root/lib

echo "Building raylib"
git submodule init
git submodule update
cd raylib/src
#make PLATFORM=PLATFORM_DRM RAYLIB_LIBTYPE=SHARED
make RAYLIB_LIBTYPE=SHARED
cd ../..
mkdir include
cp raylib/src/libraylib.so fsverify_root/lib/libraylib.so
cp raylib/src/raylib.h include/raylib.h

echo "Building fbwarn"
cd fbwarn
autoreconf --install
#./configure CFLAGS="-O2 -std=gnu99 -DEGL_NO_X11 -DLPATFORM_DRM -I../../include -L../../fsverify_root/lib"
./configure CFLAGS="-O2 -std=gnu99 -I../../include -L../../fsverify_root/lib"
make
patchelf --replace-needed libraylib.so.500 /fsverify/lib/libraylib.so src/fbwarn
cp src/fbwarn ../fsverify_root/bin/fbwarn
cd ..
cp graphic.bvg fsverify_root/share/warn.bvg

echo "Configuring fsverify"
sed 's|FBWARNLOCATION|"/fsverify/bin/fbwarn"|g' ./verifyConfig.go > ./verifyConfig.go.tmp
sed 's|BVGLOCATION|"/fsverify/share/warn.bvg"|g' ./verifyConfig.go.tmp > verify/config/config.go

echo "Building fsverify"
cd verify
go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ../fsverify_root/bin/fsverify
gcc getScreensize.c -o ../fsverify_root/bin/getscreensize
cd ..

echo "Building verifysetup"
cd verifysetup
go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ../fsverify_root/bin/verifysetup
cd ..

echo "Done."
