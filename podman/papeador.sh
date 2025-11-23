#!/bin/sh

# timeout 3s time su - runner "submission" </tmp/test-input 2>/tmp/time.output >/tmp/output
timeout 3s time /bin/submission </tmp/test-input 2>/tmp/time.output >/tmp/output
STATUS="$?"


if [ "$STATUS" = "143" ]; then
        echo "Time limit exceeded"
        exit
elif [ "$STATUS" -ne "0" ]; then
        echo "Runtime error"
        exit
fi


diff /tmp/output /tmp/expected-output >/dev/null
if [ "$?" = "1" ]; then
        echo "Wrong answer"
        exit
fi

awk 'NR==1{print $3}' /tmp/time.output
