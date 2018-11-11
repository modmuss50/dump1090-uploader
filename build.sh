#!/usr/bin/env bash

echo 'Updating libs'
go get -u github.com/modmuss50/dump1090-uploader

rm -rf output
mkdir output

platforms=("windows/amd64" "windows/386" "linux/amd64" "linux/386" "linux/arm" "linux/arm64" "darwin/amd64")

version="1.0.0"

if ! [[ -z "${BUILD_NUMBER}" ]]; then
  version=$version'.'${BUILD_NUMBER}
fi

for platform in "${platforms[@]}"
do
    echo 'Building for ' ${platform}
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='output/dump1090-uploader-'$GOOS'-'$GOARCH'-'$version
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done