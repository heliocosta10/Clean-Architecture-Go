# Orders — Clean Architecture em Go

## Sobre o projeto

Neste desafio, desenvolvi uma aplicação em Go para listar pedidos (*Orders*), utilizando os conceitos de Clean Architecture.

A ideia principal é utilizar o mesmo caso de uso, `ListOrders`, por meio de três interfaces diferentes: REST, gRPC e GraphQL. Dessa forma, a regra de negócio fica separada dos detalhes de comunicação e do banco de dados.

Também implementei a integração com PostgreSQL, as migrações automáticas e a configuração dos serviços com Docker Compose, para facilitar a execução do projeto.

## Tecnologias utilizadas

- Go
- PostgreSQL
- REST com `net/http`
- gRPC e Protocol Buffers
- GraphQL
- Docker e Docker Compose
- pgx para acesso ao PostgreSQL

## Como executar o projeto

Para executar a aplicação, preciso apenas do Docker Desktop com Docker Compose disponível.

Na raiz do projeto, executo:

```bash
docker compose up
```

Esse comando inicia o PostgreSQL e a aplicação. O Compose verifica a disponibilidade do banco antes de iniciar o serviço dependente, e a aplicação também possui tentativas de conexão caso o banco ainda não esteja pronto.

Depois de estabelecer a conexão, a aplicação executa automaticamente as migrações pendentes antes de disponibilizar os serviços REST, GraphQL e gRPC.

As migrações aplicadas ficam registradas na tabela `schema_migrations`, evitando a execução repetida de migrações que já foram concluídas.

Caso precise reconstruir a imagem da aplicação, posso executar:

```bash
docker compose up --build
```

Para encerrar os serviços, utilizo `Ctrl+C`. Se quiser remover os containers e também os volumes persistidos, executo:

```bash
docker compose down -v
```

**Atenção:** o comando acima remove os volumes do Compose e pode apagar os dados armazenados no PostgreSQL.

## Portas e serviços

| Serviço | Porta / endereço |
|---|---|
| REST | `http://localhost:8080/order` |
| GraphQL | `http://localhost:8080/graphql` |
| gRPC | `localhost:50051` |
| PostgreSQL | `localhost:5432` |

REST e GraphQL compartilham o servidor HTTP na porta `8080`. O gRPC utiliza a porta `50051`.

## Como testar

### 1. REST

O arquivo `api.http`, localizado na raiz do projeto, contém requisições prontas para testar a API. Posso executá-las pelo VS Code utilizando a extensão REST Client.

As principais requisições são:

- `POST http://localhost:8080/order` — cria um pedido de teste.
- `GET http://localhost:8080/order` — lista os pedidos cadastrados.

O `POST` pode ser utilizado para inserir pedidos de exemplo e facilitar a verificação da listagem.

### 2. GraphQL

Acessando o endpoint `http://localhost:8080/graphql`, posso executar uma consulta para listar os pedidos:

```graphql
query {
  ListOrders {
    id
    customer
    description
    amount
    created_at
  }
}
```

A consulta retorna os campos solicitados de cada pedido.

### 3. gRPC

O contrato do serviço está definido em `proto/orders.proto`. O serviço `orders.OrdersService` disponibiliza o método `ListOrders`.

Com o servidor em execução e o `grpcurl` instalado, posso testar a listagem pelo terminal:

```bash
grpcurl -plaintext localhost:50051 orders.OrdersService/ListOrders
```

O servidor também possui Reflection habilitada, permitindo que ferramentas compatíveis consultem a definição dos serviços disponíveis.

### 4. Testes automatizados

Para executar os testes automatizados do projeto localmente, utilizo:

```bash
go test ./...
```

Os testes do caso de uso verificam o comportamento da listagem e outros cenários da camada de negócio, incluindo o tratamento de erros do repositório e as validações relacionadas à criação de pedidos.

Além dos testes automatizados, executei verificações manuais para confirmar o funcionamento das interfaces REST, GraphQL e gRPC e a persistência dos pedidos no PostgreSQL.

## Arquitetura

A aplicação foi organizada para que as três interfaces utilizem o mesmo caso de uso de listagem:

```text
REST ─────────┐
              │
gRPC ─────────┼──> ListOrders Use Case
              │             │
GraphQL ──────┘             ▼
                     OrderRepository
                             │
                             ▼
                        PostgreSQL
```

Essa organização ajuda a manter as responsabilidades separadas e evita duplicar a regra de negócio em cada interface.

### Estrutura principal

- `internal/domain`: contém a entidade de domínio `Order`.
- `internal/usecase`: contém os casos de uso e o contrato do repositório.
- `internal/adapter/postgres`: implementa o acesso ao PostgreSQL utilizando pgx.
- `internal/adapter/rest`: contém o adaptador HTTP REST.
- `internal/adapter/rpc`: contém o adaptador gRPC e o código Protobuf gerado.
- `internal/adapter/graphqlapi`: contém o schema e o handler GraphQL.
- `migrations`: contém os arquivos de migração do banco de dados.
- `proto`: contém a definição do serviço gRPC.

## Resultado

Com este projeto, consegui trabalhar com Clean Architecture em Go e expor um mesmo caso de uso por meio de REST, gRPC e GraphQL.

Também pratiquei a integração com PostgreSQL, o gerenciamento de migrações, a organização das camadas e a inicialização dos serviços com Docker Compose.

O objetivo foi manter a aplicação organizada, facilitar sua execução e demonstrar que diferentes interfaces podem compartilhar a mesma regra de negócio.
