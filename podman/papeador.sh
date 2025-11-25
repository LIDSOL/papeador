#!/bin/sh

TIMECMD='time'

which time >/dev/null
STATUS="$?"

if [ "$STATUS" -ne "0" ]; then
    TIMECMD='busybox time'
fi

TIMELIMIT="$(cat /tmp/timelimit.txt)s"

mkdir -p /tmp/my-outputs

for file in $(ls /tmp/inputs); do
    timeout $TIMELIMIT $TIMECMD su -c "submission" runner </tmp/inputs/$file 2>/tmp/time.output >/tmp/my-outputs/$file
    STATUS="$?"


    if [ "$STATUS" = "143" ]; then
	    echo "Time limit exceeded"
	    break
    elif [ "$STATUS" -ne "0" ]; then
	    echo "Runtime error"
	    break
    fi


    diff /tmp/my-outputs/$file /tmp/expected-outputs/$file >/dev/null
    if [ "$?" = "1" ]; then
	    echo "Wrong answer"
	    break
    fi

    awk 'NR==1{print $3}' /tmp/time.output
done

echo "done" >/tmp/done

