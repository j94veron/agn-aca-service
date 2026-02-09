FROM golang:1.24-bookworm

WORKDIR /go/src/app-agn

# Instala dependencias del sistema (incluye libaio1)
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl unzip libaio1 && \
    rm -rf /var/lib/apt/lists/*

# Copia los módulos y descarga dependencias de Go
COPY go.mod go.sum ./
RUN go mod download

# Copia el código fuente
COPY . .

# Variables de entorno para Oracle
ENV ORACLE_HOME=/opt/oracle/instantclient_21_12
ENV PATH=$PATH:$ORACLE_HOME
ENV LD_LIBRARY_PATH=$ORACLE_HOME
ENV CONFIG_ENV=local

# Compila la aplicación
RUN go build -o /app-agn ./cmd/api

# Crea un entrypoint simple dentro de la imagen (sin archivo externo)
RUN printf '%s\n' \
  '#!/bin/bash' \
  'set -e' \
  'mkdir -p /opt/oracle/instantclient_21_12/lib' \
  'if [ -f /opt/oracle/instantclient_21_12/libclntsh.so ]; then' \
  '  ln -sf /opt/oracle/instantclient_21_12/libclntsh.so /opt/oracle/instantclient_21_12/lib/libclntsh.so' \
  'fi' \
  'exec /app-agn' \
  > /entrypoint.sh && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]