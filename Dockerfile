# Usa a imagem oficial do Go como base
FROM golang:1.20-alpine

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia o arquivo go.mod e go.sum (se existir) para o diretório de trabalho
COPY go.mod ./

# Baixa as dependências do Go e gera o arquivo go.sum
RUN go mod tidy

# Copia o código fonte para o diretório de trabalho
COPY . .

# Compila o aplicativo Go
RUN go build -o /app/main . && chmod +x /app/main

# Comando para rodar o aplicativo
CMD ["/app/main"]