# private-docker-registry-delete-images

O dockerfile deste projeto vai gerar uma imagem composta por 2 componentes principais:

1. Programa em GO que pode ser executado para apagar imagens antigas.
1. Uma cópia do [docker registry oficial](https://hub.docker.com/_/registry).

## Como apagar imagens antigas
1. Acessar o container do registry: `docker exec -it $CONTAINER_ID sh`
2. Executar o programa que apaga as imagens: `./delete-images/delete-images`
3. Seguir as instruções do programa, selecionando os repositórios e as tags que deseja-se apagar.

## Executando o GC
1. Executar os passos anteriores e selecionar a opção 999 do menu do programa `delete-images`


<br>
<br>
<br>
---

# Documentação sobre remover imagens antigas do registry

## Leituras obrigatórias:
- [Garbage Collection](https://docs.docker.com/registry/garbage-collection/#run-garbage-collection)
- [Docker Registry HTTP API V2](https://docs.docker.com/registry/spec/api/)
- [ROADMAP sobre delete de imagens](https://github.com/docker/distribution/blob/master/ROADMAP.md#deletes)

**OBS:** Para ser possível deletar usando a API o registry deve ter iniciado com a variável de ambiente `REGISTRY_STORAGE_DELETE_ENABLED=true`. [Mais informações](https://docs.docker.com/registry/configuration/#delete)

---

## Como remover uma imagem do registry:
1 - Buscar a lista de repositórios disponíveis:

| Método | URL |
| ------ | ------ |
| GET | `https://YOUR_REGISTRY_URL_HERE/v2/_catalog `|

2 - Selecionar um dos repositórios e listar as tags disponíveis:

| Método | URL |
| ------ | ------ |
| GET | `https://YOUR_REGISTRY_URL_HERE/v2/$NOME_REPOSITORIO/tags/list` |

3.  Escolher a tag que deseja deletar e buscar seu manifesto:

| Método | URL | HEADERS |
| ------ | ------ | ------ |
| GET | `https://YOUR_REGISTRY_URL_HERE/v2/$NOME_REPOSITORIO/manifests/$TAG` |Accept: application/vnd.docker.distribution.manifest.v2+json

4.  Na resposta do get da etapa 3 haverá um header chamado `Docker-Content-Digest` este valor é o que deve ser usado para fazer o delete da imagem/tag.

5.  Realizar o delete da tag:

| Método | URL |
| ------ | ------ |
| DELETE | `https://YOUR_REGISTRY_URL_HERE/v2/$NOME_REPOSITORIO/manifests/$DOCKER_CONTENT_DIGEST` |

**Exemplo da url de delete**:  https://YOUR_REGISTRY_URL_HERE/v2/dpe/agenda-web/manifests/sha256:0165d13b01778b74a95963df2bab396985e9590e80f9b62cdac6b4a6a0e2d16c

Se você receber o código de retorno `200` é porque a imagem foi excluída.

## Estado atual do GC
Como pode ser visto em ["ROADMAP sobre delete de imagens"](https://github.com/docker/distribution/blob/master/ROADMAP.md#deletes), o pessoal do docker tem receio em apagar coisas que não deveriam ser deletadas.

Atualmente, para apagar uma imagem é preciso excluir o arquivo de manifesto usando a API do registry como explicado acima e em seguida executar manualmente o comando do GC dentro do container: `bin/registry garbage-collect /etc/docker/registry/config.yml`

Conforme alertado na documentação o seguinte ponto precisa ser observado:
> Note: You should ensure that the registry is in read-only mode or not running at all. If you were to upload an image while garbage collection is running, there is the risk that the image’s layers are mistakenly deleted leading to a corrupted image. 