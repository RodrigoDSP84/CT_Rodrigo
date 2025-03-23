package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Função para esperar o PostgreSQL ficar pronto
func esperarPostgreSQL(db *sql.DB) error {
	var err error
	for i := 0; i < 10; i++ { // Tenta 10 vezes
		err = db.Ping()
		if err == nil {
			return nil // Conexão bem-sucedida
		}
		log.Printf("Tentativa %d: PostgreSQL não está pronto. Aguardando...", i+1)
		time.Sleep(5 * time.Second) // Espera 5 segundos antes de tentar novamente
	}
	return fmt.Errorf("não foi possível conectar ao PostgreSQL após várias tentativas: %v", err)
}

// Função para validar o tamanho do CPF/CNPJ
func validarCPFCNPJ(cpfCnpj string) bool {
	// Remove caracteres não numéricos
	cpfCnpj = regexp.MustCompile(`\D`).ReplaceAllString(cpfCnpj, "")

	// Valida CPF
	if len(cpfCnpj) == 11 {
		return validarCPF(cpfCnpj)
	}

	// Valida CNPJ
	if len(cpfCnpj) == 14 {
		return validarCNPJ(cpfCnpj)
	}

	return false
}

// Função para validar os digitos do CPF
func validarCPF(cpf string) bool {
	cpf = regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return false
	}

	// Cálculo dos dígitos verificadores
	var soma int
	for i := 0; i < 9; i++ {
		soma += int(cpf[i]-'0') * (10 - i)
	}
	resto := soma % 11
	digito1 := 0
	if resto >= 2 {
		digito1 = 11 - resto
	}

	soma = 0
	for i := 0; i < 10; i++ {
		soma += int(cpf[i]-'0') * (11 - i)
	}
	resto = soma % 11
	digito2 := 0
	if resto >= 2 {
		digito2 = 11 - resto
	}

	return cpf[9] == byte(digito1+'0') && cpf[10] == byte(digito2+'0')
}

// Função para validar os digitos do CNPJ
func validarCNPJ(cnpj string) bool {
	cnpj = regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")

	if len(cnpj) != 14 {
		return false
	}

	// Cálculo dos dígitos verificadores
	var soma int
	pesos := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	for i := 0; i < 12; i++ {
		soma += int(cnpj[i]-'0') * pesos[i]
	}
	resto := soma % 11
	digito1 := 0
	if resto >= 2 {
		digito1 = 11 - resto
	}

	soma = 0
	pesos = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	for i := 0; i < 13; i++ {
		soma += int(cnpj[i]-'0') * pesos[i]
	}
	resto = soma % 11
	digito2 := 0
	if resto >= 2 {
		digito2 = 11 - resto
	}

	return cnpj[12] == byte(digito1+'0') && cnpj[13] == byte(digito2+'0')
}

// Função para higienizar texto (remover acentos e converter para maiusculas)
func higienizarTexto(texto string) string {
	texto = strings.ToUpper(texto)
	texto = strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'â', 'ã', 'ä':
			return 'A'
		case 'é', 'è', 'ê', 'ë':
			return 'E'
		case 'í', 'ì', 'î', 'ï':
			return 'I'
		case 'ó', 'ò', 'ô', 'õ', 'ö':
			return 'O'
		case 'ú', 'ù', 'û', 'ü':
			return 'U'
		case 'ç':
			return 'C'
		default:
			return r
		}
	}, texto)
	return texto
}

func main() {
	// Conecta ao banco de dados PostgreSQL
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Espera o PostgreSQL ficar pronto
	err = esperarPostgreSQL(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conectado ao banco de dados com sucesso!")

	// Cria a tabela se não existir
	_, err = db.Exec(`
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
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Abre o arquivo base_teste.txt
	file, err := os.Open("base_teste.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Inicia uma transação
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Prepara a declaração de inserção
	stmt, err := tx.Prepare(`
		INSERT INTO clientes (
			cpf, private, incompleto, data_ultima_compra, ticket_medio, ticket_ultima_compra,
			loja_mais_frequente, loja_ultima_compra, cpf_valido, cnpj_valido
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Lê o arquivo linha por linha
	scanner := bufio.NewScanner(file)
	scanner.Scan() // Ignora o cabeçalho
	for scanner.Scan() {
		linha := scanner.Text()

		// Faz o parsing da linha
		cpf := strings.TrimSpace(linha[0:18])          // CPF (posicoes 1 a 19)
		private := strings.TrimSpace(linha[19:30]) == "1" // PRIVATE (posicoes 20 a 31)
		incompleto := strings.TrimSpace(linha[31:42]) == "1" // INCOMPLETO (posicoes 32 a 43)
		dataUltimaCompraStr := strings.TrimSpace(linha[43:64]) // DATA DA ULTIMA COMPRA (posicoes 44 a 65)
		ticketMedioStr := strings.TrimSpace(linha[65:86]) // TICKET MEDIO (posicoes 66 a 87)
		ticketUltimaCompraStr := strings.TrimSpace(linha[87:110]) // TICKET DA ULTIMA COMPRA (posicoes 88 a 111)
		lojaMaisFrequente := strings.TrimSpace(linha[111:130]) // LOJA MAIS FREQUENTE (posicoes 112 a 131)
		lojaUltimaCompra := strings.TrimSpace(linha[131:]) // LOJA DA ULTIMA COMPRA (posicoes 132 ao fim do arquivo)

		// Higieniza os dados
		cpf = higienizarTexto(cpf)
		lojaMaisFrequente = higienizarTexto(lojaMaisFrequente)
		lojaUltimaCompra = higienizarTexto(lojaUltimaCompra)

		// Valida CPF/CNPJ
		cpfValido := validarCPFCNPJ(cpf)
		cnpjValido := validarCPFCNPJ(lojaMaisFrequente) && validarCPFCNPJ(lojaUltimaCompra)

		// Trata a data_ultima_compra
		var dataUltimaCompra sql.NullTime
		if dataUltimaCompraStr != "NULL" && dataUltimaCompraStr != "" && dataUltimaCompraStr != "0" {
			// Converte a string para time.Time
			data, err := time.Parse("2006-01-02", dataUltimaCompraStr)
			if err != nil {
				log.Printf("Erro ao converter data: %v. Valor será definido como NULL.", err)
				dataUltimaCompra = sql.NullTime{Valid: false}
			} else {
				dataUltimaCompra = sql.NullTime{Time: data, Valid: true}
			}
		} else {
			dataUltimaCompra = sql.NullTime{Valid: false}
		}

		// Trata os campos numericos (ticket_medio e ticket_ultima_compra)
		var ticketMedio sql.NullFloat64
		if ticketMedioStr != "NULL" && ticketMedioStr != "" && ticketMedioStr != "NU" {
			valor, err := strconv.ParseFloat(strings.Replace(ticketMedioStr, ",", ".", -1), 64)
			if err != nil {
				log.Printf("Erro ao converter ticket_medio: %v. Valor será definido como NULL.", err)
				ticketMedio = sql.NullFloat64{Valid: false}
			} else {
				ticketMedio = sql.NullFloat64{Float64: valor, Valid: true}
			}
		} else {
			ticketMedio = sql.NullFloat64{Valid: false}
		}

		var ticketUltimaCompra sql.NullFloat64
		if ticketUltimaCompraStr != "NULL" && ticketUltimaCompraStr != "" && ticketUltimaCompraStr != "NU" {
			valor, err := strconv.ParseFloat(strings.Replace(ticketUltimaCompraStr, ",", ".", -1), 64)
			if err != nil {
				log.Printf("Erro ao converter ticket_ultima_compra: %v. Valor será definido como NULL.", err)
				ticketUltimaCompra = sql.NullFloat64{Valid: false}
			} else {
				ticketUltimaCompra = sql.NullFloat64{Float64: valor, Valid: true}
			}
		} else {
			ticketUltimaCompra = sql.NullFloat64{Valid: false}
		}

		// Insere os dados na transação
		_, err = stmt.Exec(cpf, private, incompleto, dataUltimaCompra, ticketMedio, ticketUltimaCompra,
			lojaMaisFrequente, lojaUltimaCompra, cpfValido, cnpjValido)
		if err != nil {
			tx.Rollback() // Desfaz a transação em caso de erro
			log.Fatal(err)
		}
	}

	// Finaliza a transação
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Dados carregados com sucesso!")
}