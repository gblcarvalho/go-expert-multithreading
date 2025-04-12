package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ViaCEPResp struct {
	CEP         string `json:"cep,omitempty"`
	Logradouro  string `json:"logradouro,omitempty"`
	Complemento string `json:"complemento,omitempty"`
	Unidade     string `json:"unidade,omitempty"`
	Bairro      string `json:"bairro,omitempty"`
	Localidade  string `json:"localidade,omitempty"`
	UF          string `json:"uf,omitempty"`
	Estado      string `json:"estado,omitempty"`
	Regiao      string `json:"regiao,omitempty"`
	IBGE        string `json:"ibge,omitempty"`
	GIA         string `json:"gia,omitempty"`
	DDD         string `json:"ddd,omitempty"`
	SIAFI       string `json:"siafi,omitempty"`
}

type BrasilAPIResp struct {
	CEP          string `json:"cep,omitempty"`
	State        string `json:"state,omitempty"`
	City         string `json:"city,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	Street       string `json:"street,omitempty"`
	Service      string `json:"service,omitempty"`
}

type CEPAPIResp struct {
	CEP          string
	State        string
	City         string
	Neighborhood string
	Street       string
	APIService   string
}

type CEPDecoderFunc func(body io.ReadCloser) (*CEPAPIResp, error)

func RequestCEPAPI(url string, decoder CEPDecoderFunc) (*CEPAPIResp, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requisição falhou com status: %s", resp.Status)
	}
	return decoder(resp.Body)
}

func ViaCEPRequest(out chan<- *CEPAPIResp, errCh chan<- error, cep string) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := RequestCEPAPI(url, func(body io.ReadCloser) (*CEPAPIResp, error) {

		var result ViaCEPResp
		if err := json.NewDecoder(body).Decode(&result); err != nil {
			return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
		}

		return &CEPAPIResp{
			CEP:          result.CEP,
			State:        result.Estado,
			City:         result.Localidade,
			Neighborhood: result.Bairro,
			Street:       result.Logradouro,
			APIService:   "viacep.com.br",
		}, nil
	})

	if err != nil {
		errCh <- err
		return
	}
	out <- resp
}

func BrasilAPIRequest(out chan<- *CEPAPIResp, errCh chan<- error, cep string) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	resp, err := RequestCEPAPI(url, func(body io.ReadCloser) (*CEPAPIResp, error) {

		var result BrasilAPIResp
		if err := json.NewDecoder(body).Decode(&result); err != nil {
			return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
		}

		return &CEPAPIResp{
			CEP:          result.CEP,
			State:        result.State,
			City:         result.City,
			Neighborhood: result.Neighborhood,
			Street:       result.Street,
			APIService:   "brasilapi.com.br",
		}, nil
	})

	if err != nil {
		errCh <- err
		return
	}
	out <- resp
}

func PrintCEPAPIResp(c *CEPAPIResp) {
	fmt.Printf("CEP ..........: %s \n", c.CEP)
	fmt.Printf("State ........: %s \n", c.State)
	fmt.Printf("City .........: %s \n", c.City)
	fmt.Printf("Neighborhood .: %s \n", c.Neighborhood)
	fmt.Printf("Street .......: %s \n", c.Street)
	fmt.Printf("API Service ..: %s \n", c.APIService)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Use: go run main.go <CEP>")
		os.Exit(1)
	}

	cep := os.Args[1]
	viaCEPChannel := make(chan *CEPAPIResp)
	brasilAPIChannel := make(chan *CEPAPIResp)
	apiErrChannel := make(chan error)

	go ViaCEPRequest(viaCEPChannel, apiErrChannel, cep)
	go BrasilAPIRequest(brasilAPIChannel, apiErrChannel, cep)

	select {
	case msg := <-viaCEPChannel:
		fmt.Println("*** Via CEP ***")
		PrintCEPAPIResp(msg)
	case msg := <-brasilAPIChannel:
		fmt.Println("*** Brasil API ***")
		PrintCEPAPIResp(msg)
	case msg := <-apiErrChannel:
		fmt.Printf("Erro: %s", msg)
	case <-time.After(time.Second):
		fmt.Print("timeout")
	}
}
