# Semana 3 - Instrumentação

Projeto da semana: utilizando o projeto referência no GitHub, façam um fork e adicionem _alguma_ instrumentação, documentando o processo e compartilhando pontos que foram descobertos no caminho (que problemas resolveram? algum ponto causou surpresa?).

## Progresso de objetivos
- ⏳  Implementar instrumentação manual usando o otel-sdk no app users
    - A configuração já está feita, mas ainda faltam alguns ajustes de logs, traces e métricas de todos os endpoints
- ⏳  Replicar a instrumentação para o restante dos apps
- ❌ Gerar testes de carga
    - Não iniciado, pode ser um sh simples usando o otel-cli
- ❌ Comparar as instrumentações e validar o nível de detalhe cada tipo
    - Não iniciado
- ⏳ Mapear possíveis problemas em nossas apps
    - Está em progresso, estou colocando no fim do documento

## Dúvidas a serem discutidas
- Hoje temos um arquivo de configuração das nossas aplicações, o otel-sdk também utiliza um arquivo de configuração. A melhor estratégia é ter dois arquivos de configuração ou usar o arquivo de configuração do otel dentro da configuração do app?
- Até que ponto vale a pena entrar em detalhes de traces e logs em operações simples como listUsers?


## Como executar localmente 
### Todas apps com o minikube

Não cheguei a testar se tudo está funcionando perfeitamente, mas os containers ficam no ar.

Requisitos
- minikube
- kubectl

```terminal
# Inicia o cluster
minikube start --driver=docker

# Usar o daemon do minikube para subir a imagem
eval $(minikube docker-env)

# Gera o binário alterado e faz o build da imagem, esteja dentro do diretorio /cmd/<app>
GOARCH=arm64 GOOS=linux go build -o users
docker build -t projeto-otel-na-pratica/users:local .
GOARCH=arm64 GOOS=linux go build -o plans
docker build -t projeto-otel-na-pratica/plans:local .
GOARCH=arm64 GOOS=linux go build -o subscriptions
docker build -t projeto-otel-na-pratica/subscriptions:local .
GOARCH=arm64 GOOS=linux go build -o payments
docker build -t projeto-otel-na-pratica/payments:local .

kubectl apply -f deployments/kubernetes.yaml

# Executa o otel-tui para a visualização dos dados
kubectl run otel-tui-temp --rm -it --image=ymtdzzz/otel-tui:latest --labels app=otel-tui
```

### Execução por app

Requisitos
- Instale o [otel-tui](https://github.com/ymtdzzz/otel-tui?tab=readme-ov-file#homebrew)

```
go run ./cmd/users/
otel-tui
curl localhost:8080/users
```

# Apresentação
Draft como objetivo de organizar uma timeline da nossa discussão na sexta

## Problemas corrigidos

## Problemas Visualizados
- A falta de testes unitários dificulta a instrumentação, principalmente em refatoração de funções que exigem propagação de contexto.
- TargetPort errados no deployments, impossibilitando a execução local de todos os apps.
- Compilação de imagens não suportam arm64, [issue](https://github.com/dosedetelemetria/projeto-otel-na-pratica/issues/13).

## Surpresas
- O container de autoinstrumentação na versão `v0.19.0-alpha` não foi possível ser adicionado. Um erro na versão do kernel impactou a lib do go-instrumentation, onde é usada operações em ebpf. Mais detalhes podem ser encontrados no [PR](https://github.com/open-telemetry/opentelemetry-go-instrumentation/pull/1567). Ainda não foi feita release da versão `v0.20.0-alpha`, mais detalhes [aqui](https://github.com/open-telemetry/opentelemetry-go-instrumentation/milestone/17). Para resolver isso basicamente setamos a imagem superior para subir o minikube, com a flag `--base-image gcr.io/k8s-minikube/minikube-iso:linux-v5`. Ao tentar colocar direto na imagem tivemos o problema abaixo:
    ```
    {"time":"2025-01-16T02:39:26.65205717Z","level":"ERROR","source":{"function":"main.main","file":"/app/cli/main.go","line":125},"msg":"failed to load instrumentation","error":"offset not found: go.opentelemetry.io/otel/trace.SpanContext:traceID (1.33.0)\noffset not found: go.opentelemetry.io/otel/trace.SpanContext:spanID (1.33.0)\noffset not found: go.opentelemetry.io/otel/trace.SpanContext:traceFlags (1.33.0)"}
    ```
    O erro é solucionável provavelmente com ajuste de versão das dependências, porém não há tempo hábil.
