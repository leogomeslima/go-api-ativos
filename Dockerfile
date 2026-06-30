# Etapa 1: Construir o executável (Builder)
FROM golang:alpine AS builder

# Instalar dependências necessárias para compilação
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copiar os arquivos de módulo e baixar as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o restante do código
COPY . .

# Compilar o binário Go (otimizado e estaticamente linkado)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

# Etapa 2: Container final (Menor e mais seguro)
FROM scratch

# Copiar certificados SSL (necessários para a API comunicar com o Azure Blob Storage via HTTPS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar informações de timezone
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copiar o executável compilado do builder
COPY --from=builder /app/main /main

# Informar qual porta o container usa
EXPOSE 8080

# Comando para rodar a aplicação
ENTRYPOINT ["/main"]
