# Orders — Clean Architecture em Go

## Sobre o projeto

Neste desafio, desenvolvi uma aplicação em Go para trabalhar com pedidos (*Orders*), aplicando os conceitos de **Clean Architecture**.

A ideia principal é utilizar o mesmo caso de uso, `ListOrders`, por meio de três interfaces diferentes: REST, gRPC e GraphQL. Dessa forma, a regra de negócio fica separada dos detalhes de comunicação e do banco de dados, evitando duplicar a lógica em cada interface.

Também implementei a integração com PostgreSQL, as migrações automáticas e a configuração dos serviços com Docker Compose, para facilitar a execução do projeto em outro ambiente.

## Tecnologias utilizadas

- Go
- PostgreSQL
- REST com `net/http`
- gRPC e Protocol Buffers
- GraphQL
- Docker e Docker Compose
- `pgx` para acesso ao PostgreSQL

## Requisitos

Para executar o projeto, preciso ter instalado:

- Docker Desktop com Docker Compose.
- Git, para clonar o repositório.
- Go, caso queira executar os testes automatizados ou compilar o projeto localmente.
- `grpcurl`, para testar a interface gRPC pelo terminal.

## Como executar o projeto

### 1. Clonar o repositório

No terminal, executo:

```bash
git clone https://github.com/heliocosta10/Clean-Architecture-Go.git
cd Clean-Architecture-Go
```

### 2. Iniciar a aplicação

Na raiz do projeto, executo:

```bash
docker compose up --build
```

Esse comando constrói a imagem da aplicação, quando necessário, e inicia os serviços definidos no Docker Compose.

O Compose verifica a disponibilidade do PostgreSQL antes de iniciar o serviço dependente. A aplicação também possui tentativas de conexão caso o banco ainda não esteja pronto.

Depois de estabelecer a conexão, a aplicação executa automaticamente as migrações pendentes antes de disponibilizar os serviços REST, GraphQL e gRPC.

As migrações aplicadas ficam registradas na tabela `schema_migrations`, evitando a execução repetida de migrações que já foram concluídas.

Quando os serviços estiverem prontos, mantenho esse terminal aberto enquanto faço os testes.

### 3. Conferir os containers

Em outro terminal PowerShell, posso verificar o estado dos serviços:

```powershell
docker compose ps
```

Para acompanhar os logs da aplicação:

```powershell
docker compose logs -f app
```

## Portas e serviços

| Serviço | Endereço |
|---|---|
| REST | `http://localhost:8080/order` |
| GraphQL | `http://localhost:8080/graphql` |
| gRPC | `localhost:50051` |
| PostgreSQL | `localhost:5432` |

REST e GraphQL compartilham o servidor HTTP na porta `8080`. O gRPC utiliza a porta `50051`.

---

## Guia de testes

Realizei os testes utilizando PowerShell no Windows. Os exemplos abaixo permitem reproduzir os principais testes do projeto.

Os pedidos criados durante os testes ficam armazenados no PostgreSQL. Por isso, os IDs e a quantidade de registros podem variar conforme os dados já existentes no banco.

### 1. Testar a API REST

A API REST permite criar pedidos e listar os registros cadastrados.

#### Criar um pedido

No PowerShell, executo:

```powershell
$body = @{
    customer    = "Cliente de teste"
    description = "Pedido criado pelo teste REST"
    amount      = 149.90
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri "http://localhost:8080/order" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body
```

A resposta deve apresentar os dados do pedido criado, incluindo o identificador (`id`) e a data de criação (`created_at`).

#### Listar os pedidos

```powershell
Invoke-RestMethod `
    -Uri "http://localhost:8080/order" `
    -Method Get
```

A resposta deve conter os pedidos cadastrados no banco de dados.

Também deixei o arquivo `api.http` na raiz do projeto, com requisições prontas para executar utilizando a extensão REST Client do VS Code.

### 2. Testar o GraphQL

O endpoint do GraphQL é:

```text
http://localhost:8080/graphql
```

#### Criar um pedido

Para testar a criação de pedidos pelo GraphQL, executo a seguinte mutation:

```powershell
$body = @{
    query = 'mutation { CreateOrder(customer: "Cliente GraphQL", description: "Pedido de teste", amount: 99.9) { id customer description amount created_at } }'
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri "http://localhost:8080/graphql" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body
```

A resposta deve apresentar os campos solicitados do pedido criado.

#### Listar os pedidos

```powershell
$body = @{
    query = 'query { ListOrders { id customer description amount created_at } }'
} | ConvertTo-Json

Invoke-RestMethod `
    -Uri "http://localhost:8080/graphql" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body
```

Essa consulta retorna os pedidos cadastrados e os campos selecionados na query.

### 3. Testar o gRPC

O contrato do serviço está definido em `proto/orders.proto`.

O serviço `orders.OrdersService` disponibiliza o método `ListOrders`, e o servidor possui Reflection habilitada, permitindo consultar os serviços disponíveis.

#### Instalar o grpcurl

Se ainda não tiver o `grpcurl` instalado e o Go estiver configurado no computador, executo:

```powershell
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Depois, verifico se o comando está disponível:

```powershell
grpcurl -version
```

Caso o comando não seja reconhecido, verifico se a pasta `bin` do diretório configurado em `GOPATH` está incluída na variável de ambiente `PATH`.

#### Listar os serviços disponíveis

```powershell
grpcurl -plaintext localhost:50051 list
```

Para consultar os métodos do serviço Orders:

```powershell
grpcurl -plaintext localhost:50051 list orders.OrdersService
```

O método esperado é:

```text
orders.OrdersService.ListOrders
```

#### Executar o método ListOrders

Com o servidor em execução, executo:

```powershell
grpcurl -plaintext -d '{}' localhost:50051 orders.OrdersService/ListOrders
```

A resposta deve conter os pedidos existentes no PostgreSQL, incluindo seus identificadores, clientes, descrições e valores.

Nos meus testes, a listagem via gRPC retornou os pedidos criados pelas interfaces REST e GraphQL, confirmando que as três interfaces acessam os mesmos dados persistidos.

### 4. Executar os testes automatizados

Na raiz do projeto, com o Go instalado e as dependências disponíveis, executo:

```powershell
go test ./...
```

Esse comando percorre os pacotes do projeto e executa os testes automatizados existentes.

Os pacotes que não possuem testes próprios podem aparecer na saída como `[no test files]`. Isso não representa uma falha.

Nos testes realizados, o pacote `internal/usecase` executou seus testes sem erros.

### 5. Verificar a compilação

Para verificar se todos os pacotes Go compilam corretamente, executo:

```powershell
go build ./...
```

O comando terminou sem apresentar erros durante a validação do projeto.

### 6. Resultado dos testes realizados

Durante a validação do projeto a partir de um clone novo, executei os seguintes testes:

- Inicialização da aplicação e do PostgreSQL com Docker Compose.
- Aplicação automática das migrações do banco.
- Criação e listagem de pedidos pela API REST.
- Criação e listagem de pedidos pelo GraphQL.
- Listagem de pedidos pelo gRPC utilizando `grpcurl`.
- Execução dos testes automatizados com `go test ./...`.
- Verificação da compilação com `go build ./...`.

Os testes manuais das três interfaces retornaram os pedidos cadastrados, e os comandos de teste e compilação terminaram sem erros.

---

## Arquitetura

A aplicação foi organizada para que as três interfaces utilizem o mesmo caso de uso de listagem:

```text
REST ───────┐
            │
gRPC ───────┼──> ListOrders Use Case
            │             │
GraphQL ────┘             ▼
                    OrderRepository
                           │
                           ▼
                      PostgreSQL
```

Essa organização ajuda a manter as responsabilidades separadas e evita duplicar a regra de negócio em cada interface.

### Estrutura principal

- `cmd/app`: ponto de entrada da aplicação.
- `internal/domain`: contém a entidade de domínio `Order`.
- `internal/usecase`: contém os casos de uso e o contrato do repositório.
- `internal/adapter/postgres`: implementa o acesso ao PostgreSQL utilizando `pgx`.
- `internal/adapter/rest`: contém o adaptador HTTP REST.
- `internal/adapter/rpc`: contém o adaptador gRPC e o código Protobuf gerado.
- `internal/adapter/graphqlapi`: contém o schema e o handler GraphQL.
- `migrations`: contém os arquivos de migração do banco de dados.
- `proto`: contém a definição do serviço gRPC.

## Como encerrar a aplicação

Para interromper a execução iniciada com `docker compose up --build`, posso pressionar `Ctrl+C` no terminal correspondente.

Também posso parar e remover os containers executando:

```powershell
docker compose down
```

Esse comando remove os containers, mas preserva os dados nos volumes nomeados do Compose.

Se quiser remover também os volumes persistidos, executo:

```powershell
docker compose down -v
```

**Atenção:** esse último comando remove os volumes do Compose e pode apagar os dados armazenados no PostgreSQL. Utilizo-o somente quando realmente quero limpar os dados de teste.

## Resultado

Com este projeto, consegui praticar Clean Architecture em Go e expor um mesmo caso de uso por meio de REST, gRPC e GraphQL.

Também trabalhei com a integração com PostgreSQL, o gerenciamento de migrações, a organização das camadas e a inicialização dos serviços com Docker Compose.

O objetivo foi manter a aplicação organizada, facilitar sua execução e demonstrar que diferentes interfaces podem compartilhar a mesma regra de negócio.
