# Semana 3 - Instrumentação

Projeto da semana: utilizando o projeto referência no GitHub, façam um fork e adicionem _alguma_ instrumentação, documentando o processo e compartilhando pontos que foram descobertos no caminho (que problemas resolveram? algum ponto causou surpresa?).

## Progresso de objetivos
- ⏳  Implementar instrumentação manual usando o otel-sdk no app users
    - A configuração já está feita, mas ainda faltam alguns ajustes de logs, traces e métricas de todos os endpoints
- ⏳  Implementar instrumentação automática no restante dos apps
    - O operator e a instrumentação do go já estão instaladas, assim como o init-container. Falta validar porque não tá sendo exportado para o otel-tui.
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
# Inicia o cluster e instala as dependencias de auto instrumentação
minikube start
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml
kubectl apply -f https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/opentelemetry-operator.yaml
kubectl apply -f deployments/instrumentation.yaml
kubectl patch deployment opentelemetry-operator-controller-manager -n opentelemetry-operator-system --type='json' -p='[
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--enable-go-instrumentation=true"}
]'

# Gera o binário alterado e faz o build da imagem
cd cmd/users
eval $(minikube docker-env)
go build -o users # se for apple silicon GOARCH=arm64 GOOS=linux
docker build -t projeto-otel-na-pratica/users:local .

kubectl apply -f deployments/kubernetes.yaml

# Executa o otel-tui para a visualização dos dados
kubectl run otel-tui-temp --rm -it --image=ymtdzzz/otel-tui:latest --labels app=otel-tui
```

### Apple silicon

Para funcionar a instrumentação no apple silicon é necessário recompilar as imagens 
```
GOARCH=arm64 GOOS=linux o build -o users|plans|subscriptions|payments
docker build -t projeto-otel-na-pratica/users|plans|subscriptions|payments:local .
```

### App user com instrumentação manual
Foi a única app onde foi aplicada a instrumentação manual, para efetuar as alterações e testar é mais rápido, por não precisa gerar a imagem e atualizar o deployment.

Requisitos
- Instale o [otel-tui](https://github.com/ymtdzzz/otel-tui?tab=readme-ov-file#homebrew)

Passos:
- Execute o app
    ```terminal
    go run ./cmd/users/
    ```
- Em outro terminal
    ```terminal
    otel-tui
    ```
- Em outro terminal
    ```terminal
    curl localhost:8080/users
    ```

# Apresentação
Draft como objetivo de organizar uma timeline da nossa discussão na sexta

## Problemas corrigidos

## Problemas Visualizados
- A falta de testes unitários dificulta a instrumentação, principalmente em refatoração de funções que exigem propagação de contexto.
- TargetPort do users está em 8081 no deploy e o app está subindo em 8080 (isso pode ser um PR aberto).
- O goreleaser não está compilando as imagens em arm64 o que dificulta a execução no apple silicon, embora tenha o QEMU no minikube o apontamento pro binário pelo instrumentation do go é um problema.

## Surpresas
