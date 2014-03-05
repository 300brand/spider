#!/bin/sh

DIRS=`find ./* -type d`
mkdir -p .coverage

for DIR in $DIRS; do
	PROFILE=`echo "$DIR" | sed -e 's#^./##' -e 's#/#_#g'`
	PROFILE_OUT=".coverage/${PROFILE}.out"
	PROFILE_HTML=".coverage/${PROFILE}.html"
	go test -coverprofile="$PROFILE_OUT" $DIR
	if [ -f "$PROFILE_OUT" ]; then
		go tool cover -html="$PROFILE_OUT" -o="$PROFILE_HTML"
	fi
done
