# syntax=docker/dockerfile:1

# Build input program
FROM golang:1.24
WORKDIR /src
COPY ./papeador-submission.go ./main.go
RUN go build -o /bin/submission ./main.go



# Create sandbox
FROM busybox:latest

# Write file
COPY ./papeador.sh /bin/papeador

COPY ./papeador-input.txt /tmp/test-input
COPY ./papeador-output.txt /tmp/expected-output

RUN chmod 600 /tmp/test-input
RUN chmod 600 /tmp/expected-output

RUN chmod 755 /bin/papeador
COPY --from=0 /bin/submission /bin/submission
CMD ["papeador"]
