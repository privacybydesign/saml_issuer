####################
# Frontend assets
####################
FROM node:22-alpine AS frontend
WORKDIR /app

COPY package.json yarn.lock build.sh ./
RUN yarn install

COPY js/ ./js/
RUN chmod +x build.sh
RUN ./build.sh

####################
# Go build
####################
FROM golang:1.26.10 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /app/js/ ./js/

# The yivi frontend and jQuery are built by the stage above and embedded
# into the binary: fail the build if they are missing.
RUN test -s js/yivi.js && test -s js/jquery.min.js

RUN CGO_ENABLED=0 go build -trimpath -o /saml_issuer .

####################
# Runtime image
####################
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=build /saml_issuer /saml_issuer

# Declared explicitly to satisfy static scanners; the :nonroot tag already
# defaults to this same user (UID 65532).
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/saml_issuer"]
