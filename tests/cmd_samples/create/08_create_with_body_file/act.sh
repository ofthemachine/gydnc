#!/bin/bash
set -e
./gydnc init . >/dev/null 2>&1 || { echo 'init failed'; exit 1; }
export GYDNC_CONFIG=./.gydnc/config.yml

BODY_FILE_CONTENT='''This is body content from a file.
It has multiple lines.'''
echo "${BODY_FILE_CONTENT}" > ./body_content.txt

./gydnc create file_test --title "File Test Title" --body-from-file ./body_content.txt
echo "--- Content of .gydnc/file_test.g6e: ---"
cat ./.gydnc/file_test.g6e