# API de Ativos em Go

Esta é uma API simples feita em Go para receber e listar os dados do formulário de ativos.

## Como rodar a API

1. Certifique-se de ter o [Go instalado](https://go.dev/doc/install) na sua máquina.
2. Abra o terminal nesta pasta (`go-api`) e execute:

```bash
go run main.go
```

O servidor será iniciado na porta `8080`.

## Endpoints

### 1. Criar um novo ativo (Receber os dados do aplicativo)
**POST** `http://localhost:8585/ativos`

**Corpo da Requisição (JSON):**
```json
{
  "descricao": "SPLIT SPRINGER MIDEA FAB: SPRINGER MIDEA, MOD: 38CCU060535MS/42ZQA60S5, 220V , [IS102]",
  "numero_ativo": "IS102",
  "numero_ativo_2": "SIS102",
  "tag_inventario": "INV-25-102",
  "estacao": "Sede CCO",
  "descricao_centro_de_custo": "SERVICOS GERAIS"
}
```

**Exemplo com cURL (Windows PowerShell):**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/ativos" -Method Post -ContentType "application/json" -Body '{"descricao": "SPLIT SPRINGER MIDEA FAB: SPRINGER MIDEA, MOD: 38CCU060535MS/42ZQA60S5, 220V , [IS102]", "numero_ativo": "IS102", "numero_ativo_2": "SIS102", "tag_inventario": "INV-25-102", "estacao": "Sede CCO", "descricao_centro_de_custo": "SERVICOS GERAIS"}'
```

### 2. Listar ativos salvos
**GET** `http://localhost:8585/ativos`

Retorna um array JSON com todos os ativos salvos na memória da API.

**Exemplo com cURL:**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/ativos" -Method Get
```
