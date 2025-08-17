package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "github.com/alanbarros/AdoraPro/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// ===================== MODELOS =====================

// EstiloProjecao define as propriedades de formatação da projeção
// swagger:model EstiloProjecao
type EstiloProjecao struct {
	TamanhoFonte int    `json:"tamanhoFonte" bson:"tamanhoFonte"`
	CorTexto     string `json:"corTexto" bson:"corTexto"`
	CorFundo     string `json:"corFundo" bson:"corFundo"`
}

// Musica representa uma música individual
// swagger:model Musica
type Musica struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Titulo          string             `json:"titulo" bson:"titulo"`
	Autor           string             `json:"autor" bson:"autor"`
	Letra           string             `json:"letra" bson:"letra"`
	Categoria       string             `json:"categoria" bson:"categoria"`
	Tags            []string           `json:"tags" bson:"tags"`
	EstiloProjecao  EstiloProjecao     `json:"estiloProjecao" bson:"estiloProjecao"`
	DataCriacao     time.Time          `json:"dataCriacao" bson:"dataCriacao"`
	DataAtualizacao time.Time          `json:"dataAtualizacao" bson:"dataAtualizacao"`
}

// Colecao representa uma coleção de músicas
// swagger:model Colecao
type Colecao struct {
	ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Titulo          string               `json:"titulo" bson:"titulo"`
	Descricao       string               `json:"descricao" bson:"descricao"`
	Autor           string               `json:"autor" bson:"autor"`
	Musicas         []primitive.ObjectID `json:"musicas" bson:"musicas"`
	DataCriacao     time.Time            `json:"dataCriacao" bson:"dataCriacao"`
	DataAtualizacao time.Time            `json:"dataAtualizacao" bson:"dataAtualizacao"`
}

// ===================== VARIÁVEIS GLOBAIS =====================

var (
	client      *mongo.Client
	musicasCol  *mongo.Collection
	colecoesCol *mongo.Collection
)

// ===================== HANDLERS MÚSICAS =====================

// @Summary Cria uma nova música
// @Tags musicas
// @Accept json
// @Produce json
// @Param musica body Musica true "Dados da música"
// @Success 201 {object} Musica
// @Failure 400 {string} string "Erro ao decodificar JSON"
// @Failure 500 {string} string "Erro ao inserir música"
// @Router /musicas [post]
func createMusica(w http.ResponseWriter, r *http.Request) {
	var musica Musica
	if err := json.NewDecoder(r.Body).Decode(&musica); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao decodificar JSON"))
		return
	}
	musica.ID = primitive.NewObjectID()
	musica.DataCriacao = time.Now()
	musica.DataAtualizacao = musica.DataCriacao
	_, err := musicasCol.InsertOne(context.TODO(), musica)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao inserir música"))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(musica)
}

// @Summary Lista todas as músicas
// @Tags musicas
// @Produce json
// @Success 200 {array} Musica
// @Failure 500 {string} string "Erro ao buscar músicas"
// @Router /musicas [get]
func listMusicas(w http.ResponseWriter, r *http.Request) {
	cur, err := musicasCol.Find(context.TODO(), map[string]interface{}{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao buscar músicas"))
		return
	}
	defer cur.Close(context.TODO())
	var musicas []Musica
	for cur.Next(context.TODO()) {
		var m Musica
		if err := cur.Decode(&m); err == nil {
			musicas = append(musicas, m)
		}
	}
	json.NewEncoder(w).Encode(musicas)
}

// @Summary Busca uma música por ID
// @Tags musicas
// @Produce json
// @Param id path string true "ID da música"
// @Success 200 {object} Musica
// @Failure 400 {string} string "ID inválido"
// @Failure 404 {string} string "Música não encontrada"
// @Router /musicas/{id} [get]
func getMusica(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID inválido"))
		return
	}
	var musica Musica
	err = musicasCol.FindOne(context.TODO(), map[string]interface{}{"_id": id}).Decode(&musica)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Música não encontrada"))
		return
	}
	json.NewEncoder(w).Encode(musica)
}

// @Summary Atualiza uma música existente
// @Tags musicas
// @Accept json
// @Produce json
// @Param id path string true "ID da música"
// @Param musica body Musica true "Dados da música"
// @Success 200 {object} Musica
// @Failure 400 {string} string "ID inválido ou erro ao decodificar JSON"
// @Failure 500 {string} string "Erro ao atualizar música"
// @Router /musicas/{id} [put]
func updateMusica(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID inválido"))
		return
	}
	var musica Musica
	if err := json.NewDecoder(r.Body).Decode(&musica); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao decodificar JSON"))
		return
	}
	musica.DataAtualizacao = time.Now()
	update := map[string]interface{}{
		"$set": musica,
	}
	_, err = musicasCol.UpdateOne(context.TODO(), map[string]interface{}{"_id": id}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao atualizar música"))
		return
	}
	musica.ID = id
	json.NewEncoder(w).Encode(musica)
}

// @Summary Remove uma música
// @Tags musicas
// @Param id path string true "ID da música"
// @Success 204 {string} string ""
// @Failure 400 {string} string "ID inválido"
// @Failure 500 {string} string "Erro ao remover música"
// @Router /musicas/{id} [delete]
func deleteMusica(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID inválido"))
		return
	}
	_, err = musicasCol.DeleteOne(context.TODO(), map[string]interface{}{"_id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao remover música"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ===================== HANDLERS COLEÇÕES =====================

// @Summary Cria uma nova coleção
// @Tags colecoes
// @Accept json
// @Produce json
// @Param colecao body Colecao true "Dados da coleção"
// @Success 201 {object} Colecao
// @Failure 400 {string} string "Erro ao decodificar JSON"
// @Failure 500 {string} string "Erro ao inserir coleção"
// @Router /colecoes [post]
func createColecao(w http.ResponseWriter, r *http.Request) {
	var colecao Colecao
	if err := json.NewDecoder(r.Body).Decode(&colecao); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao decodificar JSON"))
		return
	}
	colecao.ID = primitive.NewObjectID()
	colecao.DataCriacao = time.Now()
	colecao.DataAtualizacao = colecao.DataCriacao
	_, err := colecoesCol.InsertOne(context.TODO(), colecao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao inserir coleção"))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(colecao)
}

// @Summary Lista todas as coleções
// @Tags colecoes
// @Produce json
// @Success 200 {array} Colecao
// @Failure 500 {string} string "Erro ao buscar coleções"
// @Router /colecoes [get]
func listColecoes(w http.ResponseWriter, r *http.Request) {
	cur, err := colecoesCol.Find(context.TODO(), map[string]interface{}{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao buscar coleções"))
		return
	}
	defer cur.Close(context.TODO())
	var colecoes []Colecao
	for cur.Next(context.TODO()) {
		var c Colecao
		if err := cur.Decode(&c); err == nil {
			colecoes = append(colecoes, c)
		}
	}
	json.NewEncoder(w).Encode(colecoes)
}

// @Summary Busca uma coleção por ID
// @Tags colecoes
// @Produce json
// @Param id path string true "ID da coleção"
// @Success 200 {object} Colecao
// @Failure 400 {string} string "ID inválido"
// @Failure 404 {string} string "Coleção não encontrada"
// @Router /colecoes/{id} [get]
func getColecao(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID inválido"))
		return
	}
	var colecao Colecao
	err = colecoesCol.FindOne(context.TODO(), map[string]interface{}{"_id": id}).Decode(&colecao)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Coleção não encontrada"))
		return
	}
	json.NewEncoder(w).Encode(colecao)
}

// @Summary Atualiza uma coleção existente
// @Tags colecoes
// @Accept json
// @Produce json
// @Param id path string true "ID da coleção"
// @Param colecao body Colecao true "Dados da coleção"
// @Success 200 {object} Colecao
// @Failure 400 {string} string "ID inválido ou erro ao decodificar JSON"
// @Failure 500 {string} string "Erro ao atualizar coleção"
// @Router /colecoes/{id} [put]
func updateColecao(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID inválido"))
		return
	}
	var colecao Colecao
	if err := json.NewDecoder(r.Body).Decode(&colecao); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao decodificar JSON"))
		return
	}
	colecao.DataAtualizacao = time.Now()
	update := map[string]interface{}{
		"$set": colecao,
	}
	_, err = colecoesCol.UpdateOne(context.TODO(), map[string]interface{}{"_id": id}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao atualizar coleção"))
		return
	}
	colecao.ID = id
	json.NewEncoder(w).Encode(colecao)
}

// @Summary Remove uma coleção
// @Tags colecoes
// @Param id path string true "ID da coleção"
// @Success 204 {string} string ""
// @Failure 400 {string} string "ID inválido"
// @Failure 500 {string} string "Erro ao remover coleção"
// @Router /colecoes/{id} [delete]
func deleteColecao(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID inválido"))
		return
	}
	_, err = colecoesCol.DeleteOne(context.TODO(), map[string]interface{}{"_id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erro ao remover coleção"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ===================== INICIALIZAÇÃO =====================

// @title Music API
// @version 1.0
// @description API REST para gerenciamento de músicas e coleções.
// @host localhost:8081
// @BasePath /
func main() {
	// Conexão com o MongoDB
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Erro ao conectar no MongoDB:", err)
	}
	db := client.Database("musicadb")
	musicasCol = db.Collection("musicas")
	colecoesCol = db.Collection("colecoes")

	r := mux.NewRouter()

	// Rotas músicas
	r.HandleFunc("/musicas", createMusica).Methods("POST")
	r.HandleFunc("/musicas", listMusicas).Methods("GET")
	r.HandleFunc("/musicas/{id}", getMusica).Methods("GET")
	r.HandleFunc("/musicas/{id}", updateMusica).Methods("PUT")
	r.HandleFunc("/musicas/{id}", deleteMusica).Methods("DELETE")

	// Rotas coleções
	r.HandleFunc("/colecoes", createColecao).Methods("POST")
	r.HandleFunc("/colecoes", listColecoes).Methods("GET")
	r.HandleFunc("/colecoes/{id}", getColecao).Methods("GET")
	r.HandleFunc("/colecoes/{id}", updateColecao).Methods("PUT")
	r.HandleFunc("/colecoes/{id}", deleteColecao).Methods("DELETE")

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Println("Servidor iniciado em :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
