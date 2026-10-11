# Fechamento automático de leilões

Descrição
--------
Projeto de exemplo (Go) para gerenciamento de leilões (auctions). O objetivo é demonstrar:
- API REST para criação, listagem e fechamento automático de leilões;
- Agendamento em background para fechar leilões com base em uma duração configurável;
- Integração com MongoDB;
- Testes BDD (godog) cobrindo fluxo de fechamento automático de leilões.

Executando a aplicação
----------------------
via Docker:

1. Na raiz do projeto, construir e rodar: 

   docker-compose up -d --build

2. A aplicação ficará disponível em: http://localhost:8080

3. Para ver logs:

   docker-compose logs -f app

4. Para parar/remover containers/volumes (após testes):

   docker-compose down

Testes (TDD e BDD)
-------------------
TDD (testes unitários / de integração com `go test`):

- Rodar todos os testes Go do repositório:

  go test ./... -v

- Rodar testes de um pacote específico (ex.):

  go test ./internal/... -v

BDD (testes comportamentais com godog):

Criado apenas um cenário para contemplar o comportamento do desafio.

Os testes BDD estão em `tests_bdd` e usam o harness `godog` via `go test` (arquivo `tests_bdd/main_test.go`). 

Há scripts para rodar todo o fluxo com Docker.

1. Tornar executável e rodar o script (ou apenas executar):

   ./tests_bdd/run_tests.sh

Opção PowerShell (Windows):

1. Abrir PowerShell na raiz do projeto e executar:

   .\tests_bdd\run_tests.ps1

O que os scripts fazem:
- sobem `docker-compose up -d --build` (app + mongodb) usando `docker-compose.yml` da raiz;
- aguardam a aplicação ficar saudável (HTTP 200 no endpoint /auction?status=0);
- executam `go test -v` dentro de `tests_bdd` (que roda o TestSuite godog);
- fazem `docker-compose down` ao final (sucesso ou falha) e exibem logs em caso de erro.

Configuração das variáveis de ambiente de tempo
-----------------------------------------------
O comportamento de agendamento/fechamento de leilões depende de variáveis de ambiente. Essas variáveis são carregadas automaticamente pelo `docker-compose` a partir de `cmd/auction/.env`. 

Principais variáveis:

- AUCTION_INTERVAL: duração do leilão (ex.: "20s", "2m"). A aplicação e testes utilizam esta variável para definir quando um leilão deve ser fechado automaticamente.
- BATCH_INSERT_INTERVAL: intervalo para inserir lotes/operações em batch (ex.: "20s").
- MAX_BATCH_SIZE: tamanho máximo do batch (inteiro, ex.: 4).

Exemplo do arquivo existente (cmd/auction/.env):
```
BATCH_INSERT_INTERVAL=20s
MAX_BATCH_SIZE=4
AUCTION_INTERVAL=20s

MONGO_INITDB_ROOT_USERNAME: admin
MONGO_INITDB_ROOT_PASSWORD: admin
MONGODB_URL=******mongodb:27017/auctions?authSource=admin
MONGODB_DB=auctions
```

Modificar as variáveis
----------------------
- Via arquivo: editar `cmd/auction/.env` com os valores desejados e (re)subir os containers com `docker-compose up -d --build`.
- Via shell (para execução local com `go run`):

  Bash / WSL / macOS:

  export AUCTION_INTERVAL="30s"
  export BATCH_INSERT_INTERVAL="15s"
  export MAX_BATCH_SIZE=8
  go run ./cmd/auction

  PowerShell (Windows):

  $env:AUCTION_INTERVAL = "30s"
  $env:BATCH_INSERT_INTERVAL = "15s"
  $env:MAX_BATCH_SIZE = "8"
  go run .\cmd\auction

Notas importantes sobre os intervalos:
- Os valores de tempo (AUCTION_INTERVAL, BATCH_INSERT_INTERVAL) devem ser strings compatíveis com `time.ParseDuration` do Go (por exemplo: "500ms", "5s", "1m30s").
- Se `AUCTION_INTERVAL` estiver ausente ou inválida, o código tem fallback para 5s em alguns pontos (ver `getAuctionDuration()` no repositório).
