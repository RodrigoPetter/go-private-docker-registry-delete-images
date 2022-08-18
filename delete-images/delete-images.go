package main

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Image struct {
	tags   []string
	data   time.Time
	digest string
	size   float64
}

func main() {

	var repositorioAtual, tagAtual = "", ""

	for repositorioAtual != "Sair" {

		tagAtual = ""

		repositories := append([]interface{}{"Sair"}, getRepositories()...)

		for index, element := range repositories {
			fmt.Println(index, " - ", element)
		}

		fmt.Println("998 - ", "Scan all repositories size (this task can take several minutes)")
		fmt.Println("999 - ", "Run Garbage Collection")

		fmt.Print("Selecione um repositório:")
		var readValue int
		_, _ = fmt.Scanf("%d", &readValue)

		if readValue == 999 {
			cmd := exec.Command("bin/registry", "garbage-collect", "--delete-untagged", "/etc/docker/registry/config.yml")
			stdout, err := cmd.Output()
			if err != nil {
				panic(err)
			}
			fmt.Println(string(stdout))
		} else if readValue == 998 {

			type repoSize struct {
				index     int
				name      string
				size      float64
				tagsCount int
			}

			repoSizes := make([]repoSize, 0)

			for index, repo := range repositories {
				//skip do sair & scan & garbage collection
				if index == 0 || index == 998 || index == 999 {
					continue
				}

				fmt.Println(repo.(string))

				tagsList := getTags(repo.(string))

				var repoSizeAccumulator float64 = 0

				//lista que guarda os digest que já foi verificado o tamanho para não contar duas vezes o mesmo tamanho de tags que compartilham os mesmos blobs
				digests := make([]string, 0)
				for _, tag := range tagsList {
					tagDigest := getManifest(repo.(string), tag.(string), false)

					if !contains(digests, tagDigest) {
						fmt.Printf("[%d] Fetching [%s] repository size... \n", index, repo.(string))
						repoSizeAccumulator += getTagSize(repo.(string), tag.(string))
						digests = append(digests, tagDigest)
					}
				}

				repoSizes = append(repoSizes, repoSize{index: index, name: repo.(string), size: repoSizeAccumulator, tagsCount: len(tagsList)})
			}

			//ordena pelo tamanho
			sort.Slice(repoSizes, func(i, j int) bool {
				return repoSizes[i].size > repoSizes[j].size
			})

			fmt.Println("Tamanho aproximado utilizado pelas imagens compactadas no registry (obs: esse tamanho não está considerando que a mesma layer pode ser compartilhada entre múltiplas imagens do docker):")

			var total float64 = 0
			for _, element := range repoSizes {
				fmt.Printf("%10.2fMB - %3d tags - %s (%d)\n", element.size, element.tagsCount, element.name, element.index)
				total += element.size
			}
			fmt.Printf("\nTotal: %7.3fGB \n", total/float64(1024))
			fmt.Println("Press any key to continue...")
			fmt.Scanln()

		} else {

			repositorioAtual = repositories[readValue].(string)

			fmt.Println(repositorioAtual)

			for tagAtual != "Voltar" && repositorioAtual != "Sair" {

				response := getTags(repositorioAtual)

				if len(response) <= 0 {
					fmt.Println("Nenhuma tag encontrada...")
					break
				}

				imagesSlice := make([]Image, 0)
				for i := 0; i < len(response); i += 1 {

					var responseTag string = response[i].(string)
					var digest = getManifest(repositorioAtual, responseTag, true)
					var found = false

					//Adiciona a tag na lista se já existir o mesmo digest
					for index, v := range imagesSlice {
						if v.digest == digest {
							imagesSlice[index].tags = append(v.tags, responseTag)
							found = true
							break
						}
					}

					//Se não existe digeest na lista cria um novo elemento
					if !found {
						imagesSlice = append(imagesSlice, Image{tags: []string{responseTag}, data: getTagData(repositorioAtual, response[i].(string)), digest: digest, size: getTagSize(repositorioAtual, responseTag)})
					}
				}

				//ordena por data
				sort.Slice(imagesSlice, func(i, j int) bool { return imagesSlice[i].data.Before(imagesSlice[j].data) })

				imagesSlice = append([]Image{{tags: []string{"Voltar"}, data: time.Now(), digest: "", size: 0}}, imagesSlice...)

				for index, element := range imagesSlice {
					fmt.Printf("%-3d - %-45s |  %s  | %7.2fMB  | %s \n", index, strings.Join(element.tags, ", "), element.data.Format("02/01/2006 15:04:05 MST"), element.size, element.digest)
				}

				fmt.Println("Selecione uma tag para exclusão:")
				var readValue string
				_, _ = fmt.Scanf("%s", &readValue)

				fmt.Println(readValue)

				for _, value := range strings.Split(readValue, ",") {

					intValue, _ := strconv.Atoi(value)
					tagAtual = imagesSlice[intValue].tags[0]

					if tagAtual != "Voltar" {
						fmt.Println("=== DELETANDO TAG: ", tagAtual, " ===")
						imageDigest := imagesSlice[intValue].digest
						fmt.Println("Status do delete: ", deleteDigest(repositorioAtual, imageDigest), "\n")
						time.Sleep(1000)
					}

				}
			}
		}

	}

}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
