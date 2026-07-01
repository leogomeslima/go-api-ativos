package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Coordenadas struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Srid      string   `json:"srid"`
}

type Ativo struct {
	ID                     string      `json:"id"`
	NumeroAtivo            string      `json:"numero_ativo"`
	NumeroAtivo2           string      `json:"numero_ativo2"`
	TagInventario          string      `json:"tag_inventario"`
	Descricao              string      `json:"descricao"`
	Classe                 string      `json:"classe"`
	Subclasse              string      `json:"subclasse"`
	CentroCusto            string      `json:"centro_custo"`
	DescricaoCentroDeCusto string      `json:"descricao_centro_custo"`
	DataColocacaoServico   string      `json:"data_colocacao_servico"`
	Custo                  float64     `json:"custo"`
	LocalizacaoOrigem      string      `json:"localizacao_origem"`
	EstacaoIdentificada    *string     `json:"estacao_identificada"`
	Coordenadas            Coordenadas `json:"coordenadas"`
	Georreferenciado       bool        `json:"georreferenciado"`
	FotoUrl                *string     `json:"foto_url"`
}

var (
	ativos        []Ativo
	mu            sync.Mutex
	blobName      = "ativos.json"
	containerName = "inventarios"
)

// getBlobClient cria o cliente de blob usando a connection string do Azure
func getBlobClient() (*azblob.Client, error) {
	connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connStr == "" {
		return nil, fmt.Errorf("AZURE_STORAGE_CONNECTION_STRING não configurada")
	}

	client, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// carregarDados faz o download do JSON do Azure Blob Storage
func carregarDados() {
	client, err := getBlobClient()
	if err != nil {
		log.Println("Aviso: Azure Blob Storage não configurado. Rodando apenas em memória temporária.")
		ativos = []Ativo{}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.DownloadStream(ctx, containerName, blobName, nil)
	if err != nil {
		// Se o blob não existir, inicializamos vazio sem erro fatal
		log.Println("Arquivo não encontrado no blob ou erro de leitura. Inicializando lista vazia.")
		ativos = []Ativo{}
		return
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&ativos); err != nil {
		log.Println("Erro ao converter JSON do Blob:", err)
		ativos = []Ativo{}
	} else {
		log.Printf("Foram carregados %d ativos do Azure Blob Storage.\n", len(ativos))
	}
}

// salvarDados faz o upload do JSON para o Azure Blob Storage
func salvarDados() {
	client, err := getBlobClient()
	if err != nil {
		log.Println("Aviso: Dados NÃO salvos na Azure. AZURE_STORAGE_CONNECTION_STRING ausente.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Converte a lista atualizada para JSON com indentação
	jsonData, err := json.MarshalIndent(ativos, "", "  ")
	if err != nil {
		log.Println("Erro ao gerar JSON para salvar:", err)
		return
	}

	// Faz upload para o contêiner (sobrescreve o arquivo no Azure)
	_, err = client.UploadBuffer(ctx, containerName, blobName, jsonData, &azblob.UploadBufferOptions{})
	if err != nil {
		log.Println("Erro ao salvar no Azure Blob Storage:", err)
	} else {
		log.Println("Lista salva no Azure Blob Storage com sucesso!")
	}
}

func main() {
	// Pega o nome do container da variável de ambiente, se existir
	if envContainer := os.Getenv("AZURE_STORAGE_CONTAINER_NAME"); envContainer != "" {
		containerName = envContainer
	}

	carregarDados()

	http.HandleFunc("/ativos", ativosHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor rodando na porta :%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func ativosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		defer mu.Unlock()

		if ativos == nil {
			ativos = []Ativo{}
		}

		if err := json.NewEncoder(w).Encode(ativos); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case http.MethodPost:
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Erro ao ler o corpo", http.StatusBadRequest)
			return
		}

		var novoAtivo Ativo
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&novoAtivo); err != nil {
			http.Error(w, "Falha ao processar o JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		atualizado := false
		for i, ativoExistente := range ativos {
			if ativoExistente.ID == novoAtivo.ID {
				ativos[i] = novoAtivo
				atualizado = true
				break
			}
		}

		if !atualizado {
			ativos = append(ativos, novoAtivo)
		}

		salvarDados()
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(novoAtivo)

	case http.MethodDelete:
		idParaDeletar := r.URL.Query().Get("id")
		if idParaDeletar == "" {
			http.Error(w, "O parâmetro 'id' é obrigatório na URL. Ex: /ativos?id=123", http.StatusBadRequest)
			return
		}

		mu.Lock()
		deletado := false
		for i, ativoExistente := range ativos {
			if ativoExistente.ID == idParaDeletar {
				// Remove o elemento da fatia
				ativos = append(ativos[:i], ativos[i+1:]...)
				deletado = true
				break
			}
		}

		if deletado {
			salvarDados()
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"mensagem": "Ativo deletado com sucesso"}`))
		} else {
			mu.Unlock()
			http.Error(w, "Ativo não encontrado", http.StatusNotFound)
		}

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}
