# Music API

API REST em Go para gerenciamento de músicas, utilizando MongoDB e Docker Compose.

## Como executar

1. **Build e subir os containers:**
   ```sh
   docker-compose up --build
   ```

2. Acesse a API em: http://localhost:8080

## Estrutura do Projeto
- `main.go`: ponto de entrada da aplicação
- `internal/handler`: handlers das rotas
- `internal/model`: modelos de dados
- `internal/repository`: acesso ao banco de dados

## Variáveis de ambiente
- `MONGO_URI`: string de conexão com o MongoDB (já configurada no docker-compose)

## Copilot
Veja `.github/copilot-instructions.md` para instruções de uso do Copilot neste projeto.
