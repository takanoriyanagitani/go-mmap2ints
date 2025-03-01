#!/bin/sh

input=./sample.d/input.dat

geninput(){
	echo generating input file...

	mkdir -p sample.d

	printf \
		'\0\0\0\0\0\0\0\0''\0\0\0\0\0\0\0\1' |
		cat > "${input}"
}

test -f "${input}" || geninput

export ENV_INTS_FILENAME="${input}"

./mmap2qwords
