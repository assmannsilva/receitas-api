# API de receitas

API REST em Go para cadastrar receitas culinárias.

## Executar

Requer Go 1.25 ou superior. Não há dependências externas.

```sh
go run .
```

A API escuta em `:8080`. Variáveis opcionais:

- `ADDR`: endereço HTTP, por exemplo `127.0.0.1:8080`.
- `DATA_FILE`: arquivo de persistência; padrão `data/recipes.json`.

## Endpoints

| Método | Rota | Resultado |
| --- | --- | --- |
| GET | `/health` | Estado do serviço |
| POST | `/recipes` | Cria receita (201 e cabeçalho Location) |
| GET | `/recipes` | Lista todas as receitas |
| GET | `/recipes/{id}` | Consulta receita |
| PUT | `/recipes/{id}` | Substitui todos os campos editáveis |
| DELETE | `/recipes/{id}` | Exclui receita (204) |

Envie `Content-Type: application/json` para POST e PUT. Exemplo:

```sh
curl -i http://localhost:8080/recipes \
  -H 'Content-Type: application/json' \
  -d '{"title":"Bolo de cenoura","ingredients":["2 cenouras","2 ovos"],"instructions":"Misture e asse por 40 minutos.","prep_minutes":50,"servings":8}'

curl http://localhost:8080/recipes
```

Use o `id` retornado para consultar, atualizar ou excluir:

```sh
curl http://localhost:8080/recipes/ID
curl -X DELETE http://localhost:8080/recipes/ID
```

Título (até 200 bytes), ingredientes e instruções são obrigatórios. O tempo de preparo em minutos e o número de porções devem ser positivos. Campos desconhecidos, JSON inválido e corpos acima de 1 MiB recebem 400; IDs inexistentes recebem 404. Erros da aplicação usam `{"error":"mensagem"}`. As respostas padrão do roteador, como método não permitido, são texto.

## Verificar a saúde da API

Com o servidor em execução:

```sh
curl http://localhost:8080/health
```

A resposta esperada é `200 OK` com o corpo `{"status":"ok"}`.

## Testes e compilação

```sh
go test -race -cover ./...
go vet ./...
go build -o bin/receitas-api .
```

## Persistência e limites

O arquivo JSON é substituído atomicamente, com sincronização de acesso entre goroutines. Os testes cobrem o ciclo CRUD, recarga do arquivo, validação, falhas de escrita e concorrência. Execute apenas uma instância por arquivo de dados; não há coordenação entre processos.

Esta versão é uma base para desenvolvimento local, sem autenticação, paginação ou TLS. Para publicar na internet, configure autenticação e HTTPS e considere um banco de dados. O arquivo de dados não entra no Git.

## Publicar no GitHub

Crie um repositório vazio no GitHub e execute dentro desta pasta:

```sh
git remote add origin https://github.com/SEU_USUARIO/receitas-api.git
git push -u origin main
```
