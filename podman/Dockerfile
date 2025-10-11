# syntax=docker/dockerfile:1

# Build input program
FROM golang:1.24
WORKDIR /src
COPY /tmp/papeador-submission.go ./main.go
RUN go build -o /bin/submission ./main.go


# Create sandbox
FROM busybox:latest

# Write file
COPY <<EOF /bin/papeador
#!/bin/sh

timeout 3s time su - runner "submission" </tmp/test-input 2>/tmp/time.output >/tmp/output
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

awk 'NR==1{print $3}' time.output
EOF

COPY /tmp/papeador-input.txt /tmp/test-input
COPY /tmp/papeador-output.txt /tmp/expected-output

RUN chmod 600 /tmp/test-input
RUN chmod 600 /tmp/expected-output

RUN chmod 755 /bin/papeador
RUN adduser runner << EOF
    correct horse battery staple
    correct horse battery staple
EOF
COPY --from=0 /bin/submission /bin/submission
CMD ["papeador"]