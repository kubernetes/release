#!/bin/bash
set -e

# Create a directory to build the image content in
rm -rf build
BUILD_DIR=build

#trap "rm -rf $BUILD_DIR" EXIT

# Create necessary sub-directorctories
mkdir -p "$BUILD_DIR/layer/Files/Windows/System32/config"

# Create the files that ProcessBaseLayer on Windows validates when unpacking container images
touch "$BUILD_DIR/layer/Files/Windows/System32/config/DEFAULT"
touch "$BUILD_DIR/layer/Files/Windows/System32/config/SAM"
touch "$BUILD_DIR/layer/Files/Windows/System32/config/SECURITY"
touch "$BUILD_DIR/layer/Files/Windows/System32/config/SOFTWARE"
touch "$BUILD_DIR/layer/Files/Windows/System32/config/SYSTEM"

# Add CC0 license to image
cp cc0-license.txt "$BUILD_DIR/layer/Files/License.txt"
cp cc0-legalcode.txt "$BUILD_DIR/layer/Files/cc0-legalcode.txt"

# Create layer.tar
echo "Creating $BUILD_DIR/layer/layer.tar"
cd "$BUILD_DIR/layer"
tar -cf layer.tar Files
cd - > /dev/null

exit