# Projeto OTel na Prática

Este é o projeto que utilizamos na Especialização em OpenTelemetry no [Dose de Telemetria](https://dosedetelemetria.com). Aqui temos uma aplicação relativamente simples, mas utilizando diversos aspectos de aplicações normais, como conexões HTTP e gRPC entre si, comunicação com banco de dados, envio e recebimento de mensagens via mensageria (message queue).

A aplicação não possui nenhuma instrumentação. Nada. Durante a especialização, vamos utilizar a aplicação para aprender diversos aspectos de observabilidade, com foco em OTel.

---

## **Sumário**

- [Módulos Disponíveis](#módulos-disponíveis)
- [Configuração](#configuração)
- [Como as coisas funcionam](#como-as-coisas-funcionam)
- [Ver a telemetria localmente](#ver-a-telemetria-localmente)
- [Contribuindo](#contribuindo)
- [Licença](#licença)

---

## Módulos Disponíveis

- **`cmd/users`**:
  - **Descrição**: Este módulo contém a aplicação principal para gerenciar usuários. Ele lida com operações como criação, atualização e exclusão de usuários.

- **`cmd/payments`**:
  - **Descrição**: Este módulo é responsável pelo processamento de pagamentos. Ele gerencia transações financeiras e integrações com gateways de pagamento. Ao receber uma requisição para um novo pagamento, coloca a requisição em uma fila de mensagens. Uma rotina na mesma aplicação recebe a mensagem e processa o pagamento, armazenando em um banco de dados SQLLite.

- **`cmd/all-in-one`**:
  - **Descrição**: Este módulo combina todas as funcionalidades em uma única aplicação. Ele é útil para desenvolvimento e testes locais, permitindo executar todos os serviços em um único processo.

- **`cmd/plans`**:
  - **Descrição**: Este módulo gerencia os planos de assinatura disponíveis. Ele lida com a criação, atualização e exclusão de planos. Aceita requisições tanto em HTTP quanto gRPC.

- **`cmd/subscriptions`**:
  - **Descrição**: Este módulo gerencia as assinaturas dos usuários aos planos. Ele lida com a criação, atualização e cancelamento de assinaturas.

---

## Configuração

Por padrão, um arquivo de configuração não é necessário, especialmente ao rodar o "all-in-one". Ao fazer a aplicação rodar separadamente, a maioria dos serviços vai precisar de um arquivo de configuração específico, que segue o seguinte formato:

```yaml
# yaml-language-server: $schema=./config-schema.yaml
payments:
  subscriptions_endpoint: http://localhost:8080/subscriptions
  sqlite:
    dsn: file::memory:?cache=shared
  nats:
    endpoint: nats://localhost:4222
    subject: payment.process
    stream: payments
    consumer_name: payments

subscriptions:
  users_endpoint: http://localhost:8080/users
  plans_endpoint: http://localhost:8080/plans

plans: {}

users: {}

server:
  endpoint:
    grpc: :8081
    http: :8080
```

---

## Como as coisas funcionam

* Os serviços "plans" e "users" não tem dependências com outros serviços. O serviço "subscriptions" precisa fazer conexões com "plans" e "users", enquanto que "payments" faz uma conexão com "subscriptions".

---

## Ver a telemetria localmente

Cada binário (`cmd/all-in-one`, `users`, `plans`, `subscriptions`, `payments`) exporta trace, log e métrica por OTLP/HTTP para `http://localhost:4318`, a menos que `OTEL_EXPORTER_OTLP_ENDPOINT` esteja definido. O collector em `deployments/otelcol.yaml` recebe nessa porta, imprime o payload no stdout e encaminha para o collector externo. Sem `EXTERNAL_OTLP_ENDPOINT`, o destino é `http://external-collector:4318`. O token, se existir, vai no header `Authorization: Bearer` a partir de `EXTERNAL_OTLP_TOKEN`.

O `deployments/kubernetes.yaml` sobe o collector e as imagens publicadas de users, plans, subscriptions e payments. Essas imagens não chamam o setup de telemetria e não apontam para o serviço `otelcol`. Aplicar esse manifesto no kind ou no minikube não envia o usuário criado.

Na raiz do repositório, com o Docker rodando:

```terminal
$ docker run --rm -p 4318:4318 \
    -e EXTERNAL_OTLP_ENDPOINT \
    -e EXTERNAL_OTLP_TOKEN \
    -v "$PWD/deployments/otelcol.yaml:/etc/otelcol/config.yaml:ro" \
    otel/opentelemetry-collector-contrib:0.161.0 \
    --config=/etc/otelcol/config.yaml
```

Em outro terminal, o `all-in-one` só sobe se o NATS e a stream `payments` existirem. Os comandos estão em [cmd/all-in-one/README.md](cmd/all-in-one/README.md). Depois:

```terminal
$ go run ./cmd/all-in-one/
$ curl -X POST localhost:8080/users \
    -H 'Content-Type: application/json' \
    -d '{"id":"user-1","name":"Ada","email":"ada@example.com","address":"Rua das Flores, 100, apto 12, Sao Paulo, SP, 01310-100"}'
```

O e-mail sai no atributo `user.email` do span e no atributo `email` do log. O endereço sai no atributo `user.address` e no corpo do log, na frase `user created at ...`. O batch do SDK leva alguns segundos; a métrica sai no intervalo de 5s. O exporter `debug` do collector imprime o payload no stdout desse container.

---

## Contribuindo

Quer ajudar a melhorar este projeto? Veja como começar no arquivo [CONTRIBUTING.md](CONTRIBUTING.md). O guia explica como criar Issues, enviar Pull Requests e seguir as melhores práticas para contribuir de forma eficiente.

---

## Licença

Este projeto está licenciado sob a licença Apache v2. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.
