# Projeto de Carga de Dados com Docker e PostgreSQL

Este projeto consiste em uma aplicação em Go que carrega dados de um arquivo `base_teste.txt` para um banco de dados PostgreSQL. O ambiente é configurado usando Docker e Docker Compose, garantindo que tudo funcione de forma isolada e consistente.

---

## Pré-requisitos

Antes de começar, certifique-se de que você tem os seguintes softwares instalados:

1. **Docker**: [Guia de instalação do Docker](https://docs.docker.com/get-docker/)
2. **Docker Compose**: [Guia de instalação do Docker Compose](https://docs.docker.com/compose/install/)

---

## Estrutura do Projeto

O projeto deve ter a seguinte estrutura de arquivos:

```
CT_Rodrigo/
├── Dockerfile
├── docker-compose.yml
├── main.go
├── base_teste.txt
├── go.mod
├── go.sum
└── README.md
```

### Descrição dos arquivos:

1. **`Dockerfile`**: Define a imagem Docker para a aplicação em Go.
2. **`docker-compose.yml`**: Configura os serviços (aplicação Go e PostgreSQL) e suas dependências.
3. **`main.go`**: Código em Go que faz a carga de dados para o PostgreSQL.
4. **`base_teste.txt`**: Arquivo de dados que será carregado no banco de dados.
5. **`go.mod`**: Arquivo de dependências do Go.
6. **`go.sum`**: Arquivo gerado automaticamente pelo Go que contém checksums criptográficos (hashes) das dependências do projeto.
7. **`README.md`**: Este arquivo, com as instruções detalhadas.

---

## Passo a Passo

### 1. Instalação do Docker e Docker Compose

Siga os links abaixo para instalar o Docker e o Docker Compose no seu sistema operacional:

- **Docker**: [Instalação do Docker](https://docs.docker.com/get-docker/)
- **Docker Compose**: [Instalação do Docker Compose](https://docs.docker.com/compose/install/)

Após a instalação, verifique se tudo está funcionando corretamente:

```bash
docker --version
docker-compose --version
```

### 2. Configuração do Projeto

1. **Crie a pasta do projeto**:
   ```bash
   mkdir CT_Rodrigo
   cd CT_Rodrigo
   ```

2. **Crie os arquivos necessários (No diretório do projeto CT_Rodrigo, os arquivos já estão criados, desta forma, pode apenas baixá-los na pasta do projeto)**:
   - `Dockerfile`:
     ```Dockerfile
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
     ```

   - `docker-compose.yml`:
     ```yaml
     services:
       app:
          build: .
          container_name: go-app
          stdin_open: true
          tty: true
          depends_on:
            - db
          environment:
            - DB_HOST=db
            - DB_USER=ctrodrigo
            - DB_PASSWORD=C@s3_t3st
            - DB_NAME=ctrodrigodb
            - DB_PORT=5432
          networks:
            - mynetwork

        db:
          image: postgres:13
          container_name: postgres-db
          environment:
            POSTGRES_USER: ctrodrigo
            POSTGRES_PASSWORD: C@s3_t3st
            POSTGRES_DB: ctrodrigodb
          ports:
            - "5432:5432"
          volumes:
            - postgres_data:/var/lib/postgresql/data
          networks:
            - mynetwork

      volumes:
        postgres_data:

      networks:
        mynetwork:
          driver: bridge
     ```

   - `main.go`: Use o código em Go que você já possui.
   - `base_teste.txt`: Coloque o arquivo de dados na pasta do projeto.
   - `go.mod`: Se ainda não tiver, crie com:
     ```bash
     go mod init ct_rodrigo
     ```

     ---

## Documentação do Banco de Dados

### Estrutura da Tabela `clientes`

A tabela `clientes` é criada automaticamente pela aplicação Go quando ela é executada pela primeira vez. Aqui está a estrutura da tabela:

```sql
CREATE TABLE IF NOT EXISTS clientes (
    id SERIAL PRIMARY KEY,
    cpf TEXT,
    private BOOLEAN,
    incompleto BOOLEAN,
    data_ultima_compra DATE,
    ticket_medio NUMERIC,
    ticket_ultima_compra NUMERIC,
    loja_mais_frequente TEXT,
    loja_ultima_compra TEXT,
    cpf_valido BOOLEAN,
    cnpj_valido BOOLEAN
);
```

#### Descrição das Colunas:

| Coluna                | Tipo      | Descrição                                      |
|-----------------------|-----------|------------------------------------------------|
| `id`                  | SERIAL    | Chave primária autoincrementada.               |
| `cpf`                 | TEXT      | CPF do cliente.                                |
| `private`             | BOOLEAN   | Indica se o cliente é privado (true/false).    |
| `incompleto`          | BOOLEAN   | Indica se o registro está incompleto.          |
| `data_ultima_compra`  | DATE      | Data da última compra do cliente.              |
| `ticket_medio`        | NUMERIC   | Valor médio do ticket de compra.               |
| `ticket_ultima_compra`| NUMERIC   | Valor do ticket da última compra.              |
| `loja_mais_frequente` | TEXT      | Loja onde o cliente mais comprou.              |
| `loja_ultima_compra`  | TEXT      | Loja da última compra do cliente.              |
| `cpf_valido`          | BOOLEAN   | Indica se o CPF é válido (true/false).         |
| `cnpj_valido`         | BOOLEAN   | Indica se o CNPJ da loja é válido (true/false).|

---

### 3. Executando o Projeto

1. **Suba os containers**:
   No terminal, dentro da pasta do projeto, execute:
   ```bash
   docker-compose up --build -d
   ```

   Isso fará o build da aplicação Go e subirá o banco de dados PostgreSQL, em segundo plano (comando -d).

2. **Verifique os logs**:
   Durante a execução, você verá logs indicando que a aplicação está conectando ao banco de dados.

   Quando o processo terminar, você verá uma mensagem como a abaixo:
   ```
   [+] Running 3/3
   ✔ app                    Built                                                                                    0.0s
   ✔ Container postgres-db  Started                                                                                  0.3s
   ✔ Container go-app       Started                                                                                  0.6s

    ```

### 4. Verificando os Dados no PostgreSQL

1. **Acesse o terminal do PostgreSQL**:
   Em um novo terminal, execute:
   ```bash
   docker exec -it postgres-db psql -U ctrodrigo -d ctrodrigodb
   ```

2. **Liste as tabelas**:
   No terminal do PostgreSQL, execute:
   ```sql
   \dt
   ```

   Você verá a tabela `clientes`.

3. **Verifique os dados**:
   Para ver os primeiros 10 registros, execute:
   ```sql
   SELECT * FROM clientes LIMIT 10;
   ```

   Para contar o número total de registros, execute:
   ```sql
   SELECT COUNT(*) FROM clientes;
   ```

   Para listar os 100 últimos registros, execute:
   ```sql
   SELECT * FROM clientes ORDER BY id DESC LIMIT 100;
   ```

4. **Saia do terminal do PostgreSQL**:
   Digite:
   ```sql
   \q
   ```

### 5. Parando os Containers

Para parar os containers, execute:
```bash
docker-compose down
```

---
